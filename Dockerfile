# Docker image for the Drone ECS plugin
#
#     cd $GOPATH/src/github.com/drone-plugins/drone-ecs
#     make deps build docker

FROM alpine:3.6

# RUN apk update && apk add --no-cache bash gawk sed grep bc coreutils
RUN apk update && \
  apk add \
    ca-certificates && \
  rm -rf /var/cache/apk/*

ADD drone-ecs /bin/

ENTRYPOINT ["/bin/drone-ecs"]
