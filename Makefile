#
# Copyright (c) 2018 Dell Technologies, Inc
#
# SPDX-License-Identifier: Apache-2.0
#

.PHONY: build clean docker run
GO=CGO_ENABLED=0 GO111MODULE=on GOOS=linux go
DOCKERS=docker_edgexproxy
.PHONY: $(DOCKERS)
MICROSERVICES=edgexproxy
.PHONY: $(MICROSERVICES)
VERSION=$(shell cat ./VERSION)
GIT_SHA=$(shell git rev-parse HEAD)

prepare:

build:
	cd cmd/edgexproxy && $(GO) build  -o  $(MICROSERVICES) .
clean:
	cd cmd/edgexproxy && rm $(MICROSERVICES)
test:
	
run:
	cd cmd/edgexproxy && ./$(MICROSERVICES) init=true

docker: $(DOCKERS)
docker_edgexproxy:
		docker build \
		 --no-cache=true --rm=true \
			--label "git_sha=$(GIT_SHA)" \
				-t edgexfoundry/docker-edgex-proxy-go:$(GIT_SHA) \
				-t edgexfoundry/docker-edgex-proxy-go:$(VERSION)-dev \
				-t edgexfoundry/docker-edgex-proxy-go:latest \
				.
