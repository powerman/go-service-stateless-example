# go-service-stateless-example [![Go Report Card](https://goreportcard.com/badge/github.com/powerman/go-service-stateless-example)](https://goreportcard.com/report/github.com/powerman/go-service-stateless-example) [![CircleCI](https://circleci.com/gh/powerman/go-service-stateless-example.svg?style=svg)](https://circleci.com/gh/powerman/go-service-stateless-example) [![Coverage Status](https://coveralls.io/repos/github/powerman/go-service-stateless-example/badge.svg?branch=master)](https://coveralls.io/github/powerman/go-service-stateless-example?branch=master)

Example how to build and test stateless Go microservice with Docker,
Consul and Nginx.

## Overview

- The service is just a trivial stateless "echo" using WebSocket protocol.
- It runs in a docker container under runit supervision and write logs to
  files on docker volume.
- It also require external consul and nginx services to implement failover
  (you can run more than one container with this service, but only one
  instance of the service will be running at once, nginx will forward
  requests to running instance, and if it crash then another instance will
  be immediately started and nginx will switch to that one).

But the most interesting part is an example of how to build and test such
a service. It supports different environments, like native developer's
workstation, in a docker on developer's workstation, and in different CI.
It also supports standard testing with `go test` (and similar tools like
`goconvey`), slower but more thorough testing with `-race`, coverage and
`gometalinter`, automated integration testing and running service with all
required environment (like consul and nginx) for manual testing.

## Requirements

Docker is not required, but recommended. Without docker you'll have to
install tools like `go` and `gometalinter` on your workstation, then you
can build a service to an executable file and run standard/thorough tests,
but to run integration/manual tests you'll need to also manually setup
required environment (like consul and nginx services) on your host.

With docker you'll have to provide an external "base image" (or use one
used in this example project), which should contain all tools like `go`
and `gometalinter`. It also makes sense to include common open source
dependencies (tools and/or Go packages) of your services into this base
image - as a cache, to speedup build/test process.

If your project use docker, then your CI should provide access to docker
too (it may provide access to remote docker by network - like CircleCI
does - i.e. bind-mount support is not required within CI).

## How to build/test a service

- `./build` is supposed to build executable and/or docker image with a
  service. In trivial case it may just run `go build` (in this case you
  may not need `./build` script at all), but usually it's a bit more
  complicated and should embed current version from git and/or static
  files into executable and/or build docker image, so keeping all of this
  in a separate script is convenient.
- `go test` (and similar tools like `goconvey`) works as usually, but it
  doesn't run integration tests by default.
    - It is possible to run integration tests with `go test
      -tags=integration`, but to make it work you'll have to manually
      setup required environment (like consul and nginx services) on your
      host. To make it easier it's recommended to use docker and
      `./test_integration` (see below) instead.
- `./test` runs more thorough tests than usual `go test`, which usually
  means running `go test -race` and/or `gometalinter` and/or checking
  tests coverage.
- `./test_integration` is same as `./test` plus it runs integration tests.
  It uses `docker-compose` to start all environment required by a service
  (like consul and nginx services) and then run `./test -tags=integration`
  in separate container with project sources connected to started
  environment. It also downloads test results (like `cover.out` file) from
  the container after success.
- `docker-compose up` (or similar commands) can be used to run a service
  (with all required environment like consul and nginx services) for
  manual testing. You may need to setup some environment variables first
  (in this example service you need to set `$EXAMPLE_PORT` to define which
  port on your workstation should be used for accessing the service).
- `./ci` does nearly the same what happens in usual CI - it runs `./build`
  and `./test_integration` in a container which have all required tools
  and access to host's docker, plus it then downloads result of
  building/testing to the host (or just use bind-mount instead). This
  script may be useful in case you don't have all required tools installed
  on your workstation, so you can't just run `./build` or `./test`.

You can also find here examples of how to configure different CI services,
but main idea is to have CI just run `./build` and `./test_integration` in
your own docker "base image" - these scripts will do most of work, which
let you have trivial CI config and easily move from one CI to another.
