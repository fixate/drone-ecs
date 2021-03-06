[![Build Status](https://drone.seattleslow.com/api/badges/josmo/drone-ecs/status.svg)](https://drone.seattleslow.com/josmo/drone-ecs)
# drone-ecs


Drone plugin to deploy or update a project on AWS ECS. For the usage information and a listing of the available options please take a look at [the docs](DOCS.md).

## Binary

Build the binary using `make`:

```
make deps build
```

### Example

```
docker run --rm                          \
  -e PLUGIN_ACCESS_KEY=<key>             \
  -e PLUGIN_SECRET_KEY=<secret>          \
  -e PLUGIN_SERVICE=<service>            \  
  -e PLUGIN_DOCKER_IMAGE=<image>         \
  -v $(pwd):$(pwd)                       \
  -w $(pwd)                              \
  peloton/drone-ecs
```

## Docker

Build the container using `make`:

```
make deps docker
```

### Example
