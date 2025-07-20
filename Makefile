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
python_venv_dir := $(base_dir)/.venv
python_venv_activate := $(python_venv_dir)/bin/activate

mockery_version := 3.2.5
kernel_name := $(shell uname -s)
machine_hardware := $(shell uname -m)
ifeq ($(machine_hardware), aarch64)
	machine_hardware := arm64
endif

export SOFTHSM2_CONF ?= $(base_dir)/softhsm2.conf
TMPDIR ?= /tmp
TMPDIR := $(abspath $(TMPDIR))

maven := mvn
ifneq (, $(shell command -v mvnd 2>/dev/null))
	maven := mvnd
endif

# These should match names in Docker .env file
export FABRIC_VERSION ?= 2.5
export NODEENV_VERSION ?= 2.5
export CA_VERSION ?= 1.5

.PHONY: default
default:
	@echo 'No default target.'

.PHONY: build
build: build-node build-java

.PHONY: build-node
build-node:
	cd '$(node_dir)' && \
		npm ci && \
		npm run build && \
		rm -f fabric-gateway-dev.tgz && \
		mv $$(npm pack) fabric-gateway-dev.tgz

.PHONY: build-scenario-node
build-scenario-node: build-node
	cd '$(scenario_dir)/node' && \
		npm install @hyperledger/fabric-gateway@file:../../node/fabric-gateway-dev.tgz && \
		npm ci


.PHONY: build-java
build-java:
	cd '$(java_dir)' && \
		$(maven) -DskipTests install

.PHONY: unit-test
unit-test: generate lint unit-test-go unit-test-node unit-test-java

.PHONY: unit-test-go
unit-test-go:
	cd '$(base_dir)' && \
		go test -timeout 10s -race -coverprofile=cover.out '$(go_dir)/...'

.PHONY: unit-test-go-pkcs11
unit-test-go-pkcs11: setup-softhsm
	cd '$(base_dir)' && \
		go test -tags pkcs11 -timeout 10s -race -coverprofile=cover.out '$(go_dir)/...'

.PHONY: unit-test-node
unit-test-node: build-node
	cd '$(node_dir)' && \
		npm test

.PHONY: unit-test-java
unit-test-java:
	cd '$(java_dir)' && \
		$(maven) test jacoco:report

.PHONY: lint
lint: golangci-lint

.PHONY: install-golangci-lint
install-golangci-lint:
	curl --fail --location --show-error --silent \
		https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
		| sh -s -- -b '$(go_bin_dir)'

$(go_bin_dir)/golangci-lint:
	$(MAKE) install-golangci-lint

.PHONY: golangci-lint
golangci-lint: $(go_bin_dir)/golangci-lint
	golangci-lint run

.PHONY: scan
scan: scan-go scan-node scan-java

.PHONY: scan-go
scan-go: scan-go-osv-scanner

.PHONY: scan-go-govulncheck
scan-go-govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck -tags pkcs11 -show verbose '$(go_dir)/...'

.PHONY: scan-go-nancy
scan-go-nancy:
	go install github.com/sonatype-nexus-community/nancy@latest
	go list -json -deps '$(go_dir)/...' | nancy sleuth

.PHONY: install-osv-scanner
install-osv-scanner:
	go install github.com/google/osv-scanner/v2/cmd/osv-scanner@latest

.PHONY: scan-go-osv-scanner
scan-go-osv-scanner: install-osv-scanner
	echo "GoVersionOverride = '$$(go env GOVERSION | sed -e 's/^go//' -e 's/-.*//')'" > '$(TMPDIR)/osv-scanner.toml'
	osv-scanner scan --config='$(TMPDIR)/osv-scanner.toml' --lockfile='$(base_dir)/go.mod'

.PHONY: scan-node
scan-node: scan-node-osv-scanner

.PHONY: scan-node-npm-audit
scan-node-npm-audit:
	cd '$(node_dir)' && \
		npm audit --omit=dev

.PHONY: scan-node-osv-scanner
scan-node-osv-scanner: install-osv-scanner
	cd '$(node_dir)' && \
		npm sbom --omit=dev --package-lock-only --sbom-format cyclonedx > bom.cdx.json && \
		osv-scanner scan --lockfile=bom.cdx.json

.PHONY: scan-java
scan-java: scan-java-osv-scanner

.PHONY: scan-java-dependency-check
scan-java-dependency-check:
	cd '$(java_dir)' && \
		$(maven) dependency-check:check -P owasp

.PHONY: scan-java-osv-scanner
scan-java-osv-scanner: install-osv-scanner
	osv-scanner scan --lockfile='$(java_dir)/pom.xml' --data-source=native

.PHONY: install-mockery
install-mockery:
	curl --fail --location --show-error --silent \
		'https://github.com/vektra/mockery/releases/download/v$(mockery_version)/mockery_$(mockery_version)_$(kernel_name)_$(machine_hardware).tar.gz' \
		| tar -C '$(go_bin_dir)' -xzf - mockery

$(go_bin_dir)/mockery:
	$(MAKE) install-mockery

.PHONY: generate
generate: $(go_bin_dir)/mockery clean-generated
	cd '$(base_dir)' && mockery

.PHONY: vendor-chaincode
vendor-chaincode:
	cd '$(scenario_dir)/fixtures/chaincode/golang/basic' && \
		GO111MODULE=on go mod vendor
	cd '$(scenario_dir)/fixtures/chaincode/golang/private' && \
		GO111MODULE=on go mod vendor

.PHONY: scenario-test-go
scenario-test-go: vendor-chaincode install-fabric-ca-client setup-softhsm
	cd '$(scenario_dir)/go' && \
		go test -timeout 20m -tags pkcs11 -v -args '$(scenario_dir)/features/'

.PHONY: scenario-test-go-no-hsm
scenario-test-go-no-hsm: vendor-chaincode
	cd '$(scenario_dir)/go' && \
		go test -timeout 20m -tags pkcs11 -v --godog.tags='~@hsm' -args '$(scenario_dir)/features/'

.PHONY: scenario-test-node
scenario-test-node: vendor-chaincode build-scenario-node install-fabric-ca-client setup-softhsm
	cd '$(scenario_dir)/node' && \
		npm test

.PHONY: scenario-test-node-no-hsm
scenario-test-node-no-hsm: vendor-chaincode build-scenario-node install-fabric-ca-client
	cd '$(scenario_dir)/node' && \
		npm run test:no-hsm

.PHONY: scenario-test-java
scenario-test-java: vendor-chaincode
	cd '$(java_dir)' && \
		$(maven) -Dmaven.javadoc.skip=true -DskipUnitTests verify

.PHONY: scenario-test
scenario-test: scenario-test-go scenario-test-node scenario-test-java

.PHONY: scenario-test-no-hsm
scenario-test-no-hsm: scenario-test-go-no-hsm scenario-test-node-no-hsm scenario-test-java

.PHONY: pull-docker-images
pull-docker-images:
	for IMAGE in peer orderer baseos ccenv tools; do \
		docker pull --quiet "ghcr.io/hyperledger/fabric-$${IMAGE}:$(FABRIC_VERSION)"; \
		docker tag "ghcr.io/hyperledger/fabric-$${IMAGE}:$(FABRIC_VERSION)" "hyperledger/fabric-$${IMAGE}:$(FABRIC_VERSION)"; \
	done
	docker pull --quiet 'ghcr.io/hyperledger/fabric-nodeenv:$(NODEENV_VERSION)'
	docker tag 'ghcr.io/hyperledger/fabric-nodeenv:$(NODEENV_VERSION)' 'hyperledger/fabric-nodeenv:$(NODEENV_VERSION)'
	docker pull --quiet 'ghcr.io/hyperledger/fabric-ca:$(CA_VERSION)'
	docker tag 'ghcr.io/hyperledger/fabric-ca:$(CA_VERSION)' 'hyperledger/fabric-ca:$(CA_VERSION)'

.PHONY: install-fabric-ca-client
install-fabric-ca-client:
	go install -tags pkcs11 github.com/hyperledger/fabric-ca/cmd/fabric-ca-client@latest

.PHONY: setup-softhsm
setup-softhsm:
	mkdir -p '$(TMPDIR)/softhsm'
	echo 'directories.tokendir = $(TMPDIR)/softhsm' > '$(SOFTHSM2_CONF)'
	softhsm2-util --init-token --slot 0 --label 'ForFabric' --pin 98765432 --so-pin 1234 || true

.PHONY: generate-docs
generate-docs: $(python_venv_activate)
	. '$(python_venv_activate)' && \
		cd '$(base_dir)' && \
		python -m pip install --quiet --require-virtualenv --disable-pip-version-check --requirement requirements.txt && \
		TZ=UTC mkdocs build --strict

$(python_venv_activate):
	python -m venv '$(python_venv_dir)'

.PHONY: generate-docs-node
generate-docs-node:
	cd '$(node_dir)' && \
		npm ci && \
		npm run generate-apidoc

.PHONY: generate-docs-java
generate-docs-java:
	cd '$(java_dir)' && \
		$(maven) javadoc:javadoc

.PHONY: test
test: shellcheck unit-test scenario-test

.PHONY: all
all: test

.PHONY: clean
clean: clean-generated clean-node clean-java clean-docs

.PHONY: clean-node
clean-node:
	cd '$(node_dir)' && rm -rf node_modules

.PHONY: clean-java
clean-java:
	cd '$(java_dir)' && $(maven) clean

.PHONY: clean-generated
clean-generated:
	find '$(go_dir)' -name mocks_test.go -delete

.PHONY: clean-docs
clean-docs:
	rm -rf '$(base_dir)/site'
	rm -rf '$(node_dir)/apidocs'
	rm -rf '$(java_dir)/target/reports/apidocs'

.PHONY: shellcheck
shellcheck:
	cd '$(base_dir)' && ./scripts/shellcheck.sh

.PHONY: format
format: format-go format-node format-java

.PHONY: format-go
format-go:
	cd '$(base_dir)' && gofmt -l -s -w .

.PHONY: format-node
format-node:
	cd '$(node_dir)' && npm run format:fix

.PHONY: format-java
format-java:
	cd '$(java_dir)' && $(maven) spotless:apply
