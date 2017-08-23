#!/bin/sh

#export PLUGIN_REGION=ap-southeast-2
#export PLUGIN_ACCESS_KEY=AKIAJQKI4T2A7DAX7ERQ 
#export PLUGIN_SECRET_KEY=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
#export PLUGIN_CLUSTER=iscreen
#export PLUGIN_SERVICE=cms
#export PLUGIN_FAMILY=iscreen-cms
#export PLUGIN_TAG=v13
#export PLUGIN_PORT_MAPPINGS='container=3000'

/usr/bin/env
echo "/bin/drone-ecs $@"
date

echo "Lets go...."
/bin/drone-ecs "$@"

