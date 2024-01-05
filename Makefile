#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

base_dir := $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

go_dir := $(base_dir)/pkg
node_dir := $(base_dir)/node
java_dir := $(base_dir)/java
scenario_dir := $(base_dir)/scenario

go_bin_dir := $(shell go env GOPATH)/bin

# PEER_IMAGE_PULL is where to pull peer image from, it can be set by external env variable
# In fabric-gateway main branch it should reflect the location of the latest fabric main branch image
PEER_IMAGE_PULL ?= hyperledger-fabric.jfrog.io/fabric-peer:amd64-2.5-stable

# PEER_IMAGE_TAG is what to tag the pulled peer image as, it will also be used in docker-compose to reference the image
# In fabric-gateway main branch this version tag should correspond to the version in the forthcoming Fabric development
# branch.
export PEER_IMAGE_TAG ?= 2.5

# TWO_DIGIT_VERSION specifies which chaincode images to pull, they will be tagged to be consistent with PEER_IMAGE_TAG
# In fabric-gateway main branch it should typically be the latest released chaincode version available in dockerhub.
TWO_DIGIT_VERSION ?= 2.5

export SOFTHSM2_CONF ?= $(base_dir)/softhsm2.conf
TMPDIR ?= /tmp

.PHONEY: default
default:
	@echo 'No default target.'

.PHONEY: build
build: build-node build-java

.PHONEY: build-node
build-node:
	cd "$(node_dir)" && \
		npm install && \
		npm run build && \
		rm -f fabric-gateway-dev.tgz && \
		mv $$(npm pack) fabric-gateway-dev.tgz

.PHONEY: build-java
build-java:
	cd "$(java_dir)" && \
		mvn -DskipTests install

.PHONEY: unit-test
unit-test: generate lint unit-test-go unit-test-node unit-test-java

.PHONEY: unit-test-go
unit-test-go:
	cd '$(base_dir)' && \
		go test -timeout 10s -race -coverprofile=cover.out '$(go_dir)/...'

.PHONEY: unit-test-go-pkcs11
unit-test-go-pkcs11: setup-softhsm
	cd '$(base_dir)' && \
		go test -tags pkcs11 -timeout 10s -race -coverprofile=cover.out '$(go_dir)/...'

.PHONEY: unit-test-node
unit-test-node: build-node
	cd "$(node_dir)" && \
		npm test

.PHONEY: unit-test-java
unit-test-java:
	cd "$(java_dir)" && \
		mvn test

.PHONEY: lint
lint: staticcheck golangci-lint

.PHONEY: staticcheck
staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck -f stylish -tags=pkcs11 '$(go_dir)/...' '$(scenario_dir)/go'

.PHONEY: install-golangci-lint
install-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go_bin_dir)

$(go_bin_dir)/golangci-lint:
	$(MAKE) install-golangci-lint

.PHONEY: golangci-lint
golangci-lint: $(go_bin_dir)/golangci-lint
	golangci-lint run

.PHONEY: scan
scan: scan-go scan-node scan-java

.PHONEY: scan-go
scan-go: scan-go-govulncheck scan-go-nancy scan-go-osv-scanner

.PHONEY: scan-go-govulncheck
scan-go-govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck -tags pkcs11 "$(go_dir)/..."

.PHONEY: scan-go-nancy
scan-go-nancy:
	go install github.com/sonatype-nexus-community/nancy@latest
	go list -json -deps "$(go_dir)/..." | nancy sleuth

.PHONEY: scan-go-osv-scanner
scan-go-osv-scanner:
	go install github.com/google/osv-scanner/cmd/osv-scanner@latest
	osv-scanner --lockfile='$(base_dir)/go.mod' || [ \( $$? -gt 1 \) -a \( $$? -lt 127 \) ]

.PHONEY: scan-node
scan-node: scan-node-npm-audit scan-node-osv-scanner

.PHONEY: scan-node-npm-audit
scan-node-npm-audit:
	cd '$(node_dir)' && \
		npm install --package-lock-only && \
		npm audit --omit=dev

.PHONEY: scan-node-osv-scanner
scan-node-osv-scanner:
	go install github.com/google/osv-scanner/cmd/osv-scanner@latest
	cd '$(node_dir)' && \
		npm install && \
		npm run sbom && \
		osv-scanner --sbom=sbom.json

.PHONEY: scan-java
scan-java: scan-java-dependency-check scan-java-osv-scanner

.PHONEY: scan-java-dependency-check
scan-java-dependency-check:
	cd '$(java_dir)' && \
		mvn dependency-check:check -P owasp

.PHONEY: scan-java-osv-scanner
scan-java-osv-scanner:
	go install github.com/google/osv-scanner/cmd/osv-scanner@latest
	cd '$(java_dir)' && \
		mvn --activate-profiles sbom -DskipTests install
	osv-scanner --sbom='$(java_dir)/target/bom.json'

.PHONEY: generate
generate:
	go install go.uber.org/mock/mockgen@latest
	go generate "$(go_dir)/..."

.PHONEY: vendor-chaincode
vendor-chaincode:
	cd "$(scenario_dir)/fixtures/chaincode/golang/basic" && \
		GO111MODULE=on go mod vendor
	cd "$(scenario_dir)/fixtures/chaincode/golang/private" && \
		GO111MODULE=on go mod vendor

.PHONEY: scenario-test-go
scenario-test-go: vendor-chaincode fabric-ca-client setup-softhsm
	go install github.com/cucumber/godog/cmd/godog@v0.12
	cd $(scenario_dir)/go && \
		go test -timeout 20m -tags pkcs11 -v -args "$(scenario_dir)/features/"

.PHONEY: scenario-test-go-no-hsm
scenario-test-go-no-hsm: vendor-chaincode
	go install github.com/cucumber/godog/cmd/godog@v0.12
	cd $(scenario_dir)/go && \
		go test -timeout 20m -tags pkcs11 -v --godog.tags='~@hsm' -args "$(scenario_dir)/features/"

.PHONEY: scenario-test-node
scenario-test-node: vendor-chaincode build-node fabric-ca-client setup-softhsm
	cd "$(scenario_dir)/node" && \
		rm -rf package-lock.json node_modules && \
		npm install && \
		npm test

.PHONEY: scenario-test-node-no-hsm
scenario-test-node-no-hsm: vendor-chaincode build-node fabric-ca-client
	cd "$(scenario_dir)/node" && \
		rm -rf package-lock.json node_modules && \
		npm install && \
		npm run test:no-hsm

.PHONEY: scenario-test-java
scenario-test-java: vendor-chaincode
	cd "$(java_dir)" && \
		mvn -Dmaven.javadoc.skip=true -DskipUnitTests verify

.PHONEY: scenario-test
scenario-test: scenario-test-go scenario-test-node scenario-test-java

.PHONEY: scenario-test-no-hsm
scenario-test-no-hsm: scenario-test-go-no-hsm scenario-test-node-no-hsm scenario-test-java

.PHONEY: fabric-ca-client
fabric-ca-client:
	go install -tags pkcs11 github.com/hyperledger/fabric-ca/cmd/fabric-ca-client@latest

.PHONEY: setup-softhsm
setup-softhsm:
	mkdir -p "$(TMPDIR)/softhsm"
	echo "directories.tokendir = $(TMPDIR)/softhsm" > "$(SOFTHSM2_CONF)"
	softhsm2-util --init-token --slot 0 --label 'ForFabric' --pin 98765432 --so-pin 1234 || true

.PHONEY: generate-docs-node
generate-docs-node:
	cd "$(node_dir)" && \
		npm install && \
		npm run generate-apidoc

.PHONEY: generate-docs-java
generate-docs-java:
	cd "$(java_dir)" && \
		mvn javadoc:javadoc

.PHONEY: test
test: shellcheck unit-test scenario-test

.PHONEY: all
all: test

.PHONEY: pull-latest-peer
pull-latest-peer:
	docker pull $(PEER_IMAGE_PULL)
	docker tag $(PEER_IMAGE_PULL) hyperledger/fabric-peer:$(PEER_IMAGE_TAG)
	# also need to retag the following images for the chaincode builder
	for IMAGE in baseos ccenv javaenv nodeenv; do \
		docker pull hyperledger/fabric-$${IMAGE}:$(TWO_DIGIT_VERSION); \
		docker tag hyperledger/fabric-$${IMAGE}:$(TWO_DIGIT_VERSION) hyperledger/fabric-$${IMAGE}:$(PEER_IMAGE_TAG); \
	done

.PHONEY: clean
clean: clean-generated clean-node clean-java

.PHONEY: clean-node
clean-node:
	rm -rf "$(node_dir)/package-lock.json" "$(node_dir)/node_modules"

.PHONEY: clean-java
clean-java:
	cd "$(java_dir)" && mvn clean

.PHONEY: clean-generated
clean-generated:
	find "$(go_dir)" -name '*_mock_test.go' -delete

.PHONEY: shellcheck
shellcheck:
	cd "$(base_dir)" && ./scripts/shellcheck.sh
