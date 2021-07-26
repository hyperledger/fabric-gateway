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
samples_dir := $(base_dir)/samples

# PEER_IMAGE_PULL is where to pull peer image from, it can be set by external env variable
PEER_IMAGE_PULL ?= hyperledger-fabric.jfrog.io/fabric-peer:amd64-latest

# PEER_IMAGE_TAG is what to tag the pulled peer image as, it will also be used in docker-compose to reference the image
PEER_IMAGE_TAG ?= 2.4

# TWO_DIGIT_VERSION specifies which chaincode images to pull, they will be tagged to be consistent with PEER_IMAGE_TAG
TWO_DIGIT_VERSION ?= 2.3

PKGNAME = github.com/hyperledger/fabric-gateway
ARCH=$(shell go env GOARCH)
MARCH=$(shell go env GOOS)-$(shell go env GOARCH)

GO_VER = 1.15.6
GO_TAGS ?=

build: build-protos build-go build-node

fabric_protos_commit = ef707fbbf8bf90dccb2808b6f1543af3068c1caa
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
	cd $(node_dir); npm install; npm run build; rm -f fabric-gateway-dev.tgz; mv $$(npm pack) fabric-gateway-dev.tgz

build-java: build-protos
	cd $(java_dir); mvn install -DskipTests

unit-test: generate unit-test-go unit-test-node unit-test-java

unit-test-go: lint staticcheck
	go test -timeout 10s -coverprofile=$(base_dir)/cover.out $(base_dir)/pkg/...

unit-test-node: build-node
	cd $(node_dir); npm test

unit-test-java: build-protos
	cd $(java_dir); mvn test

lint:
	golint -set_exit_status $(base_dir)/pkg/... $(scenario_dir)/go

staticcheck:
	staticcheck $(base_dir)/pkg/... $(scenario_dir)/go

sample-network: pull-latest-peer vendor-chaincode
	cd $(scenario_dir)/go; GATEWAY_NO_SHUTDOWN=TRUE godog $(scenario_dir)/features/transactions.feature

sample-network-clean:
	docker ps -aq | xargs docker rm -f; docker images -q 'dev-*' | xargs docker rmi -f; docker network prune --force

run-samples: | sample-network run-samples-go run-samples-node run-samples-java sample-network-clean

run-samples-go:
	cd $(samples_dir)/go; go run sample.go

run-samples-node: build-node
	cd $(samples_dir)/node; rm -f package-lock.json; rm -rf node_modules; npm install; npm run build; npm start

run-samples-java: build-java
	cd $(samples_dir)/java; mvn test

generate:
	go generate ./pkg/...

vendor-chaincode:
	cd $(scenario_dir)/fixtures/chaincode/golang/basic; GO111MODULE=on go mod vendor
	cd $(scenario_dir)/fixtures/chaincode/golang/private; GO111MODULE=on go mod vendor

scenario-test-go: vendor-chaincode
	cd $(scenario_dir)/go; godog --tags="~@hsm" $(scenario_dir)/features/

scenario-test-node: vendor-chaincode build-node
	cd $(scenario_dir)/node; rm -f package-lock.json; rm -rf node_modules; npm install; SOFTHSM2_CONF=${HOME}/softhsm2.conf npm test

scenario-test-java: vendor-chaincode build-protos
	cd $(java_dir); mvn verify

scenario-test: scenario-test-go scenario-test-node scenario-test-java

.PHONEY: generate-docs-node
generate-docs-node: build-node
	cd $(node_dir); npm run generate-apidoc

.PHONEY: generate-docs-java
generate-docs-java: build-protos
	cd $(java_dir); mvn javadoc:javadoc

test: unit-test scenario-test

all: test

pull-latest-peer:
	docker pull $(PEER_IMAGE_PULL)
	docker tag $(PEER_IMAGE_PULL) hyperledger/fabric-peer:$(PEER_IMAGE_TAG)
	# also need to retag the following images for the chaincode builder
	for IMAGE in baseos ccenv javaenv nodeenv; do \
		docker pull hyperledger/fabric-$$IMAGE:$(TWO_DIGIT_VERSION); \
		docker tag hyperledger/fabric-$$IMAGE:$(TWO_DIGIT_VERSION) hyperledger/fabric-$$IMAGE:$(PEER_IMAGE_TAG); \
	done

.PHONEY: clean
clean: clean-protos clean-generated

.PHONEY: clean-protos
clean-protos:
	-rm -rf fabric-protos
	-rm $(pb_files)

clean-generated:
	find ./pkg -name '*_mock_test.go' -delete
