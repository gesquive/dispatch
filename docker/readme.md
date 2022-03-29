# [gesquive/dispatch](https://github.com/gesquive/dispatch)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/gesquive/dispatch/blob/master/LICENSE)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/gesquive/dispatch)
[![Build Status](https://img.shields.io/circleci/build/github/gesquive/dispatch?style=flat-square)](https://circleci.com/gh/gesquive/dispatch)
[![Coverage Report](https://img.shields.io/codecov/c/gh/gesquive/dispatch?style=flat-square)](https://codecov.io/gh/gesquive/dispatch)
[![Github Release](https://img.shields.io/github/v/tag/gesquive/dispatch?style=flat-square)](https://github.com/gesquive/dispatch)

# Supported Architectures

This image supports multiple architectures:

- `amd64`, `x86-64`
- `armv7`, `armhf`
- `arm64`, `aarch64`

Docker images are uploaded with using Docker manifest lists to make multi-platform deployments easer. More info can be found from [Docker](https://github.com/docker/distribution/blob/master/docs/spec/manifest-v2-2.md#manifest-list)

You can simply pull the image using `gesquive/dispatch` and docker should retreive the correct image for your architecture.

# Supported Tags
If you want a specific version of `dispatch` you can pull it by specifying a version tag.

## Version Tags
This image provides versions that are available via tags. 

| Tag    | Description |
| ------ | ----------- |
| `latest` | Latest stable release |
| `0.9.0`  | Stable release v0.9.0 |
| `0.9.0-<git_hash>` | Development preview of version v0.9.0 at the given git hash |

# Usage

Here are some example snippets to help you get started creating a container.

## Docker CLI

```shell
docker run \
  --name=dispatch \
  -p 2525:2525 \
  -v path/to/config:/config \
  -e "DISPATCH_TARGET_AUTH_TOKEN=<AUTH_TOKEN>" \
  -e "DISPATCH_TARGET_NAME=test target" \
  -e "DISPATCH_TARGET_TO_ADDRESS=email@mail.com" \
  -e "DISPATCH_FROM_ADDRESS=test1@server.net,test2@server.net" \
  --restart unless-stopped \
  gesquive/dispatch
```

## Docker Compose
Compatible with docker-compose v2 schemas.

```docker
---
version: "2"
services:
  dispatch:
    image: gesquive/dispatch
    container_name: dispatch
    volumes:
      - path/to/config:/config
    environment:
      - DISPATCH_TARGET_AUTH_TOKEN="<AUTH_TOKEN>"
      - DISPATCH_TARGET_NAME="test target"
      - DISPATCH_TARGET_TO_ADDRESS="email@mail.com"
      - DISPATCH_FROM_ADDRESS="test1@server.net,test2@server.net"
    ports:
      - 2525:2525
    restart: unless-stopped
```

# Parameters
The container defines the following parameters that you can set:

| Parameter | Function |
| --------- | -------- |
| `-p 2525`     | The port for the dispatch REST API |
| `-v /config`  | The dispatch config goes here |
| `-e "DISPATCH_TARGET_AUTH_TOKEN=<AUTH_TOKEN>"` | The target auth token |
| `-e "DISPATCH_TARGET_NAME=test target"` | The target name |
| `-e "DISPATCH_TARGET_TO_ADDRESS=email@mail.com"` | The target too address |
| `-e "DISPATCH_FROM_ADDRESS=test1@server.net,test2@server.net"` | The target from addresses |
