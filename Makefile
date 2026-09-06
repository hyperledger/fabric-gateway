#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

BASE_DIR := $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

GO_DIR := $(BASE_DIR)/pkg
NODE_DIR := $(BASE_DIR)/node
JAVA_DIR := $(BASE_DIR)/java
SCENARIO_DIR := $(BASE_DIR)/scenario
BIN_DIR := $(BASE_DIR)/.bin

PYTHON_VENV_DIR := $(BASE_DIR)/.venv
PYTHON_VENV_ACTIVATE := $(PYTHON_VENV_DIR)/bin/activate

GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
OSV_SCANNER := $(BIN_DIR)/osv-scanner
MOCKERY := $(BIN_DIR)/mockery
FABRIC_CA_CLIENT := $(BIN_DIR)/fabric-ca-client

KERNEL_NAME := $(shell uname -s)
LOWERCASE_KERNEL_NAME := $(shell echo '$(KERNEL_NAME)' | tr '[:upper:]' '[:lower:]')

MACHINE_HARDWARE := $(shell uname -m)
ifeq ($(MACHINE_HARDWARE), aarch64)
	MACHINE_HARDWARE := arm64
endif

AMD_ARM_MACHINE_HARDWARE := $(MACHINE_HARDWARE)
ifeq ($(AMD_ARM_MACHINE_HARDWARE), x86_64)
	AMD_ARM_MACHINE_HARDWARE := amd64
endif

export SOFTHSM2_CONF ?= $(BASE_DIR)/softhsm2.conf
TMPDIR ?= /tmp
TMPDIR := $(abspath $(TMPDIR))

MAVEN := mvn
ifneq (, $(shell command -v mvnd 2>/dev/null))
	MAVEN := mvnd
endif

# If GH_TOKEN environment variable is set, use it as the GitHub API auth token
GH_API_AUTH := $(if $(GH_TOKEN),--header 'Authorization: Bearer $(GH_TOKEN)',)

OSV_SCANNER_ARGS ?=

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
	cd '$(NODE_DIR)' && \
		npm ci && \
		npm run build && \
		rm -f fabric-gateway-dev.tgz && \
		mv $$(npm pack) fabric-gateway-dev.tgz

.PHONY: build-scenario-node
build-scenario-node: build-node
	cd '$(SCENARIO_DIR)/node' && \
		npm install @hyperledger/fabric-gateway@file:../../node/fabric-gateway-dev.tgz && \
		npm ci


.PHONY: build-java
build-java:
	cd '$(JAVA_DIR)' && \
		$(MAVEN) -DskipTests install

.PHONY: unit-test
unit-test: generate lint unit-test-go unit-test-node unit-test-java

.PHONY: unit-test-go
unit-test-go:
	cd '$(BASE_DIR)' && \
		go test -timeout 10s -race -coverprofile=cover.out '$(GO_DIR)/...'

.PHONY: unit-test-go-pkcs11
unit-test-go-pkcs11: setup-softhsm
	cd '$(BASE_DIR)' && \
		go test -tags pkcs11 -timeout 10s -race -coverprofile=cover.out '$(GO_DIR)/...'

.PHONY: unit-test-node
unit-test-node: build-node
	cd '$(NODE_DIR)' && \
		npm test

.PHONY: unit-test-java
unit-test-java:
	cd '$(JAVA_DIR)' && \
		$(MAVEN) test jacoco:report

.PHONY: lint
lint: golangci-lint

.PHONY: install-golangci-lint
install-golangci-lint: uninstall-golangci-lint $(GOLANGCI_LINT)

.PHONY: uninstall-golangci-lint
uninstall-golangci-lint:
	rm -f '$(GOLANGCI_LINT)'

$(GOLANGCI_LINT):
	mkdir -p '$(dir $(GOLANGCI_LINT))'
	curl --fail --location --show-error --silent \
		https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
		| sh -s -- -b '$(dir $(GOLANGCI_LINT))'

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

.PHONY: scan
scan: scan-go scan-node scan-java

.PHONY: scan-go
scan-go: scan-go-osv-scanner

.PHONY: scan-go-govulncheck
scan-go-govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck -tags pkcs11 -show verbose '$(GO_DIR)/...'

.PHONY: scan-go-nancy
scan-go-nancy:
	go install github.com/sonatype-nexus-community/nancy@latest
	go list -json -deps '$(GO_DIR)/...' | nancy sleuth

.PHONY: install-osv-scanner
install-osv-scanner: uninstall-osv-scanner $(OSV_SCANNER)
	
.PHONY: uninstall-osv-scanner
uninstall-osv-scanner:
	rm -f '$(OSV_SCANNER)'

$(OSV_SCANNER):
	mkdir -p '$(dir $(OSV_SCANNER))'
	curl --fail --location --show-error --silent --output '$(OSV_SCANNER)' \
    	'https://github.com/google/osv-scanner/releases/latest/download/osv-scanner_$(LOWERCASE_KERNEL_NAME)_$(AMD_ARM_MACHINE_HARDWARE)'
	chmod u+x '$(OSV_SCANNER)'

.PHONY: scan-go-osv-scanner
scan-go-osv-scanner: $(OSV_SCANNER)
	echo "GoVersionOverride = '$$(go env GOVERSION | sed -e 's/^go//' -e 's/-.*//')'" > '$(TMPDIR)/osv-scanner.toml' && \
		if [ -r '$(BASE_DIR)/osv-scanner.toml' ]; then cat '$(BASE_DIR)/osv-scanner.toml' >> '$(TMPDIR)/osv-scanner.toml'; fi && \
		$(OSV_SCANNER) scan source --config='$(TMPDIR)/osv-scanner.toml' --lockfile='$(BASE_DIR)/go.mod' $(OSV_SCANNER_ARGS)

.PHONY: scan-node
scan-node: scan-node-osv-scanner

.PHONY: scan-node-npm-audit
scan-node-npm-audit:
	cd '$(NODE_DIR)' && \
		npm audit --omit=dev

.PHONY: scan-node-osv-scanner
scan-node-osv-scanner: $(OSV_SCANNER)
	cd '$(NODE_DIR)' && \
		npm sbom --omit=dev --package-lock-only --sbom-format cyclonedx > bom.cdx.json && \
		$(OSV_SCANNER) scan source --lockfile=bom.cdx.json $(OSV_SCANNER_ARGS)

.PHONY: scan-java
scan-java: scan-java-osv-scanner

.PHONY: scan-java-dependency-check
scan-java-dependency-check:
	cd '$(JAVA_DIR)' && \
		$(MAVEN) dependency-check:check -P owasp

.PHONY: scan-java-osv-scanner
scan-java-osv-scanner: $(OSV_SCANNER)
	$(OSV_SCANNER) scan source --lockfile='$(JAVA_DIR)/pom.xml' $(OSV_SCANNER_ARGS)

.PHONY: install-mockery
install-mockery: uninstall-mockery $(MOCKERY)

.PHONY: uninstall-mockery
uninstall-mockery:
	rm -f '$(MOCKERY)'

# Silent to prevent printing of auth token
.SILENT: $(MOCKERY)
$(MOCKERY):
	mkdir -p '$(dir $(MOCKERY))'
	mockery_version=$$(curl --fail --show-error --silent $(GH_API_AUTH) https://api.github.com/repos/vektra/mockery/releases | jq --raw-output '.[].tag_name' | sort --version-sort | tail -1) && \
		curl --fail --location --show-error --silent \
			"https://github.com/vektra/mockery/releases/download/$${mockery_version}/mockery_$${mockery_version#v}_$(KERNEL_NAME)_$(MACHINE_HARDWARE).tar.gz" \
			| tar -C '$(dir $(MOCKERY))' -xzf - mockery
	chmod u+x '$(MOCKERY)'

.PHONY: generate
generate: $(MOCKERY) clean-generated
	cd '$(BASE_DIR)' && $(MOCKERY)

.PHONY: vendor-chaincode
vendor-chaincode:
	cd '$(SCENARIO_DIR)/fixtures/chaincode/golang/basic' && \
		GO111MODULE=on go mod vendor
	cd '$(SCENARIO_DIR)/fixtures/chaincode/golang/private' && \
		GO111MODULE=on go mod vendor

.PHONY: scenario-test-go
scenario-test-go: vendor-chaincode $(FABRIC_CA_CLIENT) setup-softhsm
	cd '$(SCENARIO_DIR)/go' && \
		go test -timeout 20m -tags pkcs11 -v -args '$(SCENARIO_DIR)/features/'

.PHONY: scenario-test-go-no-hsm
scenario-test-go-no-hsm: vendor-chaincode
	cd '$(SCENARIO_DIR)/go' && \
		go test -timeout 20m -tags pkcs11 -v --godog.tags='~@hsm' -args '$(SCENARIO_DIR)/features/'

.PHONY: scenario-test-node
scenario-test-node: vendor-chaincode build-scenario-node $(FABRIC_CA_CLIENT) setup-softhsm
	cd '$(SCENARIO_DIR)/node' && \
		npm test

.PHONY: scenario-test-node-no-hsm
scenario-test-node-no-hsm: vendor-chaincode build-scenario-node
	cd '$(SCENARIO_DIR)/node' && \
		npm run test:no-hsm

.PHONY: scenario-test-java
scenario-test-java: vendor-chaincode
	cd '$(JAVA_DIR)' && \
		$(MAVEN) -Dmaven.javadoc.skip=true -DskipUnitTests verify

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
install-fabric-ca-client: uninstall-fabric-ca-client $(FABRIC_CA_CLIENT)

.PHONY: uninstall-fabric-ca-client
uninstall-fabric-ca-client:	
	rm -f '$(FABRIC_CA_CLIENT)'

$(FABRIC_CA_CLIENT):
	go install -tags pkcs11 github.com/hyperledger/fabric-ca/cmd/fabric-ca-client@latest
	mkdir -p '$(dir $(FABRIC_CA_CLIENT))'
	cp -fp $(shell go env GOBIN)/fabric-ca-client $(FABRIC_CA_CLIENT)

.PHONY: setup-softhsm
setup-softhsm:
	mkdir -p '$(TMPDIR)/softhsm'
	echo 'directories.tokendir = $(TMPDIR)/softhsm' > '$(SOFTHSM2_CONF)'
	softhsm2-util --init-token --slot 0 --label 'ForFabric' --pin 98765432 --so-pin 1234 || true

.PHONY: generate-docs
generate-docs: install-docs-requirements $(PYTHON_VENV_ACTIVATE)
	. '$(PYTHON_VENV_ACTIVATE)' && \
		TZ=UTC zensical build --strict --clean

$(PYTHON_VENV_ACTIVATE):
	python -m venv '$(PYTHON_VENV_DIR)'

.PHONY: install-docs-requirements
install-docs-requirements: $(PYTHON_VENV_ACTIVATE)
	. '$(PYTHON_VENV_ACTIVATE)' && \
		cd '$(BASE_DIR)' && \
		python -m pip install --quiet --require-virtualenv --disable-pip-version-check --requirement requirements.txt

.PHONY: serve-docs
serve-docs: install-docs-requirements $(PYTHON_VENV_ACTIVATE)
	. '$(PYTHON_VENV_ACTIVATE)' && \
		cd '$(BASE_DIR)' && \
		TZ=UTC zensical serve --strict

.PHONY: generate-docs-node
generate-docs-node:
	cd '$(NODE_DIR)' && \
		npm ci && \
		npm run generate-apidoc

.PHONY: generate-docs-java
generate-docs-java:
	cd '$(JAVA_DIR)' && \
		$(MAVEN) javadoc:javadoc

.PHONY: test
test: shellcheck unit-test scenario-test

.PHONY: all
all: test

.PHONY: clean
clean: clean-generated clean-node clean-java clean-docs

.PHONY: clean-node
clean-node:
	rm -rf '$(NODE_DIR)/node_modules'

.PHONY: clean-java
clean-java:
	cd '$(JAVA_DIR)' && $(MAVEN) clean

.PHONY: clean-generated
clean-generated:
	find '$(GO_DIR)' -name mocks_test.go -delete

.PHONY: clean-docs
clean-docs:
	rm -rf '$(BASE_DIR)/site'
	rm -rf '$(NODE_DIR)/apidocs'
	rm -rf '$(JAVA_DIR)/target/reports/apidocs'

.PHONY: shellcheck
shellcheck:
	cd '$(BASE_DIR)' && ./scripts/shellcheck.sh

.PHONY: format
format: format-go format-node format-java

.PHONY: format-go
format-go:
	cd '$(BASE_DIR)' && gofmt -l -s -w .

.PHONY: format-node
format-node:
	cd '$(NODE_DIR)' && npm run format:fix
	cd '$(SCENARIO_DIR)/node' && npm run format:fix

.PHONY: format-java
format-java:
	cd '$(JAVA_DIR)' && $(MAVEN) spotless:apply
