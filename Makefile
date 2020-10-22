#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

base_dir := $(PWD)

go_dir := $(base_dir)/pkg
node_dir := $(base_dir)/node
scenario_dir := $(base_dir)/scenario

ALPINE_VER ?= 3.12
BASE_VERSION = 2.2.0
# TWO_DIGIT_VERSION is derived, e.g. "2.0", especially useful as a local tag
# for two digit references to most recent baseos and ccenv patch releases
TWO_DIGIT_VERSION = $(shell echo $(BASE_VERSION) | cut -d '.' -f 1,2)

PKGNAME = github.com/hyperledger/fabric-gateway
ARCH=$(shell go env GOARCH)
MARCH=$(shell go env GOOS)-$(shell go env GOARCH)

# defined in common/metadata/metadata.go
METADATA_VAR = Version=$(BASE_VERSION)
METADATA_VAR += CommitSHA=$(EXTRA_VERSION)
METADATA_VAR += BaseDockerLabel=$(BASE_DOCKER_LABEL)
METADATA_VAR += DockerNamespace=$(DOCKER_NS)

GO_VER = 1.14.4
GO_TAGS ?=

include docker-env.mk

build: build-go build-node

build-go:
	go build -o bin/gateway cmd/gateway/*.go

build-node:
	cd $(node_dir); npm install

unit-test: unit-test-go unit-test-node

unit-test-go:
	go test -cover $(base_dir)/pkg/... $(base_dir)/cmd/gateway

unit-test-node: build-node
	cd $(node_dir); npm test

lint:
	golint $(base_dir)/pkg/... $(base_dir)/cmd/gateway

scenario-test-go: docker
	cd $(scenario_dir)/go; godog $(scenario_dir)/features/

scenario-test-node: build-node docker
	cd $(scenario_dir)/node; npm install; npm test

scenario-test: scenario-test-go scenario-test-node

test: unit-test scenario-test

all: test

docker: build-go
	@echo "Building Docker image $(DOCKER_NS)/fabric-gateway"
	@mkdir -p $(@D)
	$(DBUILD) -f images/gateway/Dockerfile \
		--build-arg GO_VER=$(GO_VER) \
		--build-arg ALPINE_VER=$(ALPINE_VER) \
		--build-arg GO_TAGS=${GO_TAGS} \
		-t $(DOCKER_NS)/fabric-gateway ./$(BUILD_CONTEXT)
	docker tag $(DOCKER_NS)/fabric-gateway $(DOCKER_NS)/fabric-gateway:$(BASE_VERSION)
	docker tag $(DOCKER_NS)/fabric-gateway $(DOCKER_NS)/fabric-gateway:$(TWO_DIGIT_VERSION)
	docker tag $(DOCKER_NS)/fabric-gateway $(DOCKER_NS)/fabric-gateway:$(DOCKER_TAG)
