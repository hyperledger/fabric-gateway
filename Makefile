#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

base_dir := $(PWD)

go_dir := $(base_dir)/pkg
node_dir := $(base_dir)/node
java_dir := $(base_dir)/java
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

GO_VER = 1.15.6
GO_TAGS ?=

include docker-env.mk

build: build-protos build-go build-node

build-protos:
	protoc --version
	rm -rf fabric-protos
	git clone https://github.com/hyperledger/fabric-protos.git
	mkdir fabric-protos/gateway
	cp protos/gateway.proto fabric-protos/gateway
	protoc -I. -I./fabric-protos --go_out=paths=source_relative:. --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:. protos/gateway.proto

build-go: build-protos
	go build -o bin/gateway cmd/gateway/*.go

build-node: build-protos
	cd $(node_dir); npm install; npm run build

unit-test: unit-test-go unit-test-node unit-test-java

unit-test-go: build-protos
	go test -coverprofile=$(base_dir)/cover.out $(base_dir)/pkg/... $(base_dir)/cmd/gateway

unit-test-node: build-node
	cd $(node_dir); npm test

unit-test-java: build-protos
	cd $(java_dir); mvn test

lint:
	golint $(base_dir)/pkg/... $(base_dir)/cmd/gateway

vendor-chaincode:
	cd $(scenario_dir)/fixtures/chaincode/golang/basic; GO111MODULE=on go mod vendor

scenario-test-go: vendor-chaincode docker
	cd $(scenario_dir)/go; godog $(scenario_dir)/features/

scenario-test-node: vendor-chaincode docker build-node
	cd $(node_dir); rm -f fabric-gateway-dev.tgz; mv $$(npm pack) fabric-gateway-dev.tgz
	cd $(scenario_dir)/node; rm -f package-lock.json; rm -rf node_modules; npm install; npm test

scenario-test-java: vendor-chaincode docker
	cd $(java_dir); mvn verify

scenario-test: scenario-test-go scenario-test-node scenario-test-java

test: unit-test scenario-test

all: test

docker: build-protos build-go
	@echo "Building Docker image $(DOCKER_NS)/fabric-gateway"
	@mkdir -p $(@D)
	$(DBUILD) --rm -f images/gateway/Dockerfile \
		--build-arg GO_VER=$(GO_VER) \
		--build-arg ALPINE_VER=$(ALPINE_VER) \
		--build-arg GO_TAGS=${GO_TAGS} \
		-t $(DOCKER_NS)/fabric-gateway ./$(BUILD_CONTEXT)
	docker tag $(DOCKER_NS)/fabric-gateway $(DOCKER_NS)/fabric-gateway:$(BASE_VERSION)
	docker tag $(DOCKER_NS)/fabric-gateway $(DOCKER_NS)/fabric-gateway:$(TWO_DIGIT_VERSION)
	docker tag $(DOCKER_NS)/fabric-gateway $(DOCKER_NS)/fabric-gateway:$(DOCKER_TAG)
