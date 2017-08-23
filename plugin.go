package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type Plugin struct {
	Key                     string
	Secret                  string
	Region                  string
	Family                  string
	TaskRoleArn             string
	Service                 string
	ContainerName           string
	DockerImage             string
	Tag                     string
	Cluster                 string
	DeploymentConfiguration string
	PortMappings            []string
	Environment             []string
	DesiredCount            int64
	CPU                     int64
	Memory                  int64
	MemoryReservation       int64
	YamlVerified            bool
	Discreet                bool
	NoServiceUpdate         bool
}

func parsePortMap(str string) (map[string]int64, error) {
	components := strings.Split(strings.Trim(str, " "), ",")
	m := make(map[string]int64)
	for _, element := range components {
		parts := strings.SplitN(element, "=", 2)
		if len(parts) < 2 {
			return nil, errors.New(fmt.Sprintf("malformed map expression: '%s'", element))
		}
		port, err := strconv.ParseInt(parts[1], 10, 64)
		m[parts[0]] = port
		if err != nil {
			return nil, errors.New(fmt.Sprintf("port number should be an integer: '%s'", element))
		}
	}

	return m, nil
}

func (p *Plugin) Exec() error {
	fmt.Println("Starting Drone AWS ECS deployment")
	defer fmt.Println("Drone AWS ECS plugin finished.")
	awsConfig := aws.Config{}

	if len(p.Key) != 0 && len(p.Secret) != 0 {
		awsConfig.Credentials = credentials.NewStaticCredentials(p.Key, p.Secret, "")
	} else {
		fmt.Println("No AWS credentials provided")
	}
	awsConfig.Region = aws.String(p.Region)
	svc := ecs.New(session.New(&awsConfig))

	Image := p.DockerImage + ":" + p.Tag
	if len(p.ContainerName) == 0 {
		p.ContainerName = p.Family + "-container"
	}

	definition := ecs.ContainerDefinition{
		Command: []*string{},

		DnsSearchDomains:      []*string{},
		DnsServers:            []*string{},
		DockerLabels:          map[string]*string{},
		DockerSecurityOptions: []*string{},
		EntryPoint:            []*string{},
		Environment:           []*ecs.KeyValuePair{},
		Essential:             aws.Bool(true),
		ExtraHosts:            []*ecs.HostEntry{},

		Image:        aws.String(Image),
		Links:        []*string{},
		MountPoints:  []*ecs.MountPoint{},
		Name:         aws.String(p.ContainerName),
		PortMappings: []*ecs.PortMapping{},

		Ulimits: []*ecs.Ulimit{},
		//User: aws.String("String"),
		VolumesFrom: []*ecs.VolumeFrom{},
		//WorkingDirectory: aws.String("String"),
	}

	if p.CPU != 0 {
		definition.Cpu = aws.Int64(p.CPU)
	}

	if p.Memory == 0 && p.MemoryReservation == 0 {
		definition.MemoryReservation = aws.Int64(128)
	} else {
		if p.Memory != 0 {
			definition.Memory = aws.Int64(p.Memory)
		}
		if p.MemoryReservation != 0 {
			definition.MemoryReservation = aws.Int64(p.MemoryReservation)
		}
	}

	// Port mappings
	for _, portMapping := range p.PortMappings {
		parsedMappings, portMappingParseErr := parsePortMap(portMapping)
		if portMappingParseErr != nil {
			return portMappingParseErr
		}

		pair := ecs.PortMapping{
			Protocol: aws.String("TransportProtocol"),
		}

		for key, value := range parsedMappings {
			switch key {
			case "container":
				pair.ContainerPort = aws.Int64(value)
			case "host":
				pair.HostPort = aws.Int64(value)
			default:
				fmt.Println(fmt.Sprintf("WARNING: invalid portmapping key '%s'", key))
			}
		}

		definition.PortMappings = append(definition.PortMappings, &pair)
	}

	// Environment variables
	for _, envVar := range p.Environment {
		parts := strings.SplitN(envVar, "=", 2)
		pair := ecs.KeyValuePair{
			Name:  aws.String(strings.Trim(parts[0], " ")),
			Value: aws.String(strings.Trim(parts[1], " ")),
		}
		definition.Environment = append(definition.Environment, &pair)
	}
	params := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			&definition,
		},
		Family:      aws.String(p.Family),
		Volumes:     []*ecs.Volume{},
		TaskRoleArn: aws.String(p.TaskRoleArn),
	}
	resp, err := svc.RegisterTaskDefinition(params)

	if err != nil {
		return err
	}

	if p.NoServiceUpdate {
		fmt.Println("no-service-update flag is true. Not updating service to new task definition.")
		return nil
	}

	val := *(resp.TaskDefinition.TaskDefinitionArn)
	sparams := &ecs.UpdateServiceInput{
		Cluster:        aws.String(p.Cluster),
		Service:        aws.String(p.Service),
		TaskDefinition: aws.String(val),
	}

	if p.DesiredCount != 0 {
		sparams.DesiredCount = aws.Int64(p.DesiredCount)
	}

	if len(p.DeploymentConfiguration) != 0 {
		cleanedDeploymentConfiguration := strings.Trim(p.DeploymentConfiguration, " ")
		parts := strings.SplitN(cleanedDeploymentConfiguration, " ", 2)
		minimumHealthyPercent, minimumHealthyPercentError := strconv.ParseInt(parts[0], 10, 64)
		if minimumHealthyPercentError != nil {
			return minimumHealthyPercentError
		}
		maximumPercent, maximumPercentErr := strconv.ParseInt(parts[1], 10, 64)
		if maximumPercentErr != nil {
			return maximumPercentErr
		}

		sparams.DeploymentConfiguration = &ecs.DeploymentConfiguration{
			MaximumPercent:        aws.Int64(maximumPercent),
			MinimumHealthyPercent: aws.Int64(minimumHealthyPercent),
		}
	}

	sresp, serr := svc.UpdateService(sparams)

	if serr != nil {
		return serr
	}

	if !p.Discreet {
		fmt.Println(sresp)
		fmt.Println(resp)
	}

	return nil

}
