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

PEER_VERSION = 2.4
ALPINE_VER ?= 3.12
BASE_VERSION = 2.3.0
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

fabric_protos_commit = 583e2aebe6867e93bd156b32889d39ee62382037
pb_files = protos/gateway/gateway.pb.go protos/gateway/gateway_grpc.pb.go

.PHONEY: build-protos
build-protos: $(pb_files)

fabric-protos:
	git clone https://github.com/hyperledger/fabric-protos.git
	cd fabric-protos && git checkout "$(fabric_protos_commit)"

$(pb_files): fabric-protos
	protoc --version
	mkdir -p protos
	protoc -I./fabric-protos --go_out=paths=source_relative:./protos --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:./protos fabric-protos/gateway/gateway.proto

build-go:
	go build -o bin/gateway cmd/gateway/*.go

build-node: build-protos
	cd $(node_dir); npm install; npm run build

unit-test: generate unit-test-go unit-test-node unit-test-java

unit-test-go: lint staticcheck
	go test -coverprofile=$(base_dir)/cover.out $(base_dir)/pkg/... $(base_dir)/cmd/gateway

unit-test-node: build-node
	cd $(node_dir); npm test

unit-test-java: build-protos
	cd $(java_dir); mvn test

lint:
	golint -set_exit_status $(base_dir)/pkg/client/... $(base_dir)/pkg/internal/... $(base_dir)/cmd/... $(scenario_dir)/go

staticcheck:
	staticcheck $(base_dir)/pkg/client/... $(base_dir)/pkg/internal/... $(base_dir)/cmd/... $(scenario_dir)/go

generate:
	go generate ./pkg/...

vendor-chaincode:
	cd $(scenario_dir)/fixtures/chaincode/golang/basic; GO111MODULE=on go mod vendor

scenario-test-go: vendor-chaincode
	cd $(scenario_dir)/go; godog $(scenario_dir)/features/

scenario-test-node: vendor-chaincode build-node
	cd $(node_dir); rm -f fabric-gateway-dev.tgz; mv $$(npm pack) fabric-gateway-dev.tgz
	cd $(scenario_dir)/node; rm -f package-lock.json; rm -rf node_modules; npm install; npm test

scenario-test-java: vendor-chaincode build-protos
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

pull-latest-peer:
	docker pull hyperledger-fabric.jfrog.io/fabric-peer:amd64-latest
	docker tag hyperledger-fabric.jfrog.io/fabric-peer:amd64-latest hyperledger/fabric-peer:$(PEER_VERSION)
	# also need to retag the following images for the chaincode builder
	for IMAGE in baseos ccenv javaenv nodeenv; do \
		docker pull hyperledger/fabric-$$IMAGE:$(TWO_DIGIT_VERSION); \
		docker tag hyperledger/fabric-$$IMAGE:$(TWO_DIGIT_VERSION) hyperledger/fabric-$$IMAGE:$(PEER_VERSION); \
	done

.PHONEY: clean
clean: clean-protos clean-generated

.PHONEY: clean-protos
clean-protos:
	-rm -rf fabric-protos
	-rm $(pb_files)

clean-generated:
	find ./pkg -name '*_mock_test.go' -delete
