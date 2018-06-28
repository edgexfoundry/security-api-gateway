#
# Copyright (c) 2018 Dell Technologies, Inc
#
# SPDX-License-Identifier: Apache-2.0
#

.PHONY: build clean docker run
GO=CGO_ENABLED=0 go
GOCGO=CGO_ENABLED=1 go
DOCKERS=docker_edgexproxy
.PHONY: $(DOCKERS)
MICROSERVICES=edgexproxy
.PHONY: $(MICROSERVICES)
VERSION=$(shell cat ./VERSION)
GIT_SHA=$(shell git rev-parse --short HEAD)
build:
    cd core && $(GO) build  -o  $(MICROSERVICES)
clean:
    cd core && rm -f $(MICROSERVICES)
run:
    cd core && ./edgexproxy init=true
docker: $(DOCKERS)
docker_edgexproxy:
        docker build \
                --label "git_sha=$(GIT_SHA)" \
                -t edgexfoundry/docker-edgex-proxy:$(GIT_SHA) \
                -t edgexfoundry/docker-edgex-proxy:$(VERSION)-dev \
                -t edgexfoundry/docker-edgex-proxy \
                .
