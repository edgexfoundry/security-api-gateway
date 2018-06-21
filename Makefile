#
# Copyright (c) 2018 Dell Technologies, Inc
#
# SPDX-License-Identifier: Apache-2.0
#

.PHONY: build run test

build: edgexsecurity
	go build ./...

edgexsecurity:
	go build -o ./edgexsecurity

run:
	cd bin && ./security-launch.sh

test:
	go test ./...
	go vet ./...