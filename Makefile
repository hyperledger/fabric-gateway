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

# PEER_IMAGE_PULL is where to pull peer image from, it can be set by external env variable
# In fabric-gateway main branch it should reflect the location of the latest fabric main branch image
PEER_IMAGE_PULL ?= hyperledger-fabric.jfrog.io/fabric-peer:amd64-latest

# PEER_IMAGE_TAG is what to tag the pulled peer image as, it will also be used in docker-compose to reference the image
# In fabric-gateway main branch this version tag should correspond to the version in the fabric main branch
PEER_IMAGE_TAG ?= 2.5

# TWO_DIGIT_VERSION specifies which chaincode images to pull, they will be tagged to be consistent with PEER_IMAGE_TAG
# In fabric-gateway main branch it should typically be the latest released chaincode version available in dockerhub.
TWO_DIGIT_VERSION ?= 2.4

PKGNAME = github.com/hyperledger/fabric-gateway
ARCH=$(shell go env GOARCH)
MARCH=$(shell go env GOOS)-$(shell go env GOARCH)

GO_VER = 1.17.8
GO_TAGS ?=

build: build-node build-java

fabric_protos_commit = 9f95521bb870cca7b765217c80aeb600e0bd5abf
pb_files = protos/gateway/gateway.pb.go protos/gateway/gateway_grpc.pb.go

.PHONEY: build-protos
build-protos: $(pb_files)

fabric-protos:
	git clone https://github.com/hyperledger/fabric-protos.git
	cd fabric-protos && \
		git checkout "$(fabric_protos_commit)"

$(pb_files): fabric-protos

build-node: build-protos
	cd $(node_dir) && \
		npm install && \
		npm run build && \
		rm -f fabric-gateway-dev.tgz && \
		mv $$(npm pack) fabric-gateway-dev.tgz

build-java: build-protos
	cd $(java_dir) && \
		mvn install -DskipTests

unit-test: generate unit-test-go unit-test-node unit-test-java

unit-test-go: lint
	go test -timeout 10s -coverprofile=$(base_dir)/cover.out $(base_dir)/pkg/...

unit-test-go-pkcs11: lint
	SOFTHSM2_CONF=${HOME}/softhsm2.conf go test -tags pkcs11 -timeout 10s -coverprofile=$(base_dir)/cover.out $(base_dir)/pkg/...

unit-test-node: build-node
	cd $(node_dir) && \
		npm test

unit-test-java: build-protos
	cd $(java_dir) && \
		mvn test

lint:
	$(base_dir)/ci/check_gofmt.sh $(base_dir)/pkg $(scenario_dir)/go
	staticcheck -f stylish -tags="pkcs11" $(base_dir)/pkg/... $(scenario_dir)/go
	go vet -tags pkcs11 $(base_dir)/pkg/... $(scenario_dir)/go
	gosec -tags pkcs11 -exclude-generated $(base_dir)/pkg/...

scan: scan-go scan-node scan-java

scan-go:
	go list -json -deps ./pkg/... | docker run --rm --interactive sonatypecommunity/nancy:latest sleuth

scan-node:
	cd $(node_dir) && \
		npm install --package-lock-only && \
		npm audit --production

scan-java: build-protos
	cd $(java_dir) && \
		mvn dependency-check:check -P owasp

generate:
	go generate ./pkg/...

vendor-chaincode:
	cd $(scenario_dir)/fixtures/chaincode/golang/basic && \
		GO111MODULE=on go mod vendor
	cd $(scenario_dir)/fixtures/chaincode/golang/private && \
		GO111MODULE=on go mod vendor

scenario-test-go: vendor-chaincode
	cd $(scenario_dir)/go && \
		SOFTHSM2_CONF=${HOME}/softhsm2.conf go test -tags pkcs11 -v -args $(scenario_dir)/features/

scenario-test-node: vendor-chaincode build-node
	cd $(scenario_dir)/node && \
		rm -rf package-lock.json node_modules && \
		npm install && \
		SOFTHSM2_CONF=${HOME}/softhsm2.conf npm test

scenario-test-java: vendor-chaincode build-java
	cd $(java_dir) && \
		mvn verify

scenario-test: scenario-test-go scenario-test-node scenario-test-java

.PHONEY: generate-docs-node
generate-docs-node: build-node
	cd $(node_dir) && \
		npm run generate-apidoc

.PHONEY: generate-docs-java
generate-docs-java: build-protos
	cd $(java_dir) && \
		mvn javadoc:javadoc

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
