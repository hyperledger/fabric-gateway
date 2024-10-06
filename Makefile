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

export SOFTHSM2_CONF ?= $(base_dir)/softhsm2.conf
TMPDIR ?= /tmp

.PHONY: default
default:
	@echo 'No default target.'

.PHONY: build
build: build-node build-java

.PHONY: build-node
build-node:
	cd '$(node_dir)' && \
		npm install && \
		npm run build && \
		rm -f fabric-gateway-dev.tgz && \
		mv $$(npm pack) fabric-gateway-dev.tgz

.PHONY: build-java
build-java:
	cd '$(java_dir)' && \
		mvn -DskipTests install

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
		mvn test jacoco:report

.PHONY: lint
lint: staticcheck golangci-lint

.PHONY: staticcheck
staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck -f stylish -tags=pkcs11 '$(go_dir)/...' '$(scenario_dir)/go'

.PHONY: install-golangci-lint
install-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b '$(go_bin_dir)'

$(go_bin_dir)/golangci-lint:
	$(MAKE) install-golangci-lint

.PHONY: golangci-lint
golangci-lint: $(go_bin_dir)/golangci-lint
	golangci-lint run

.PHONY: scan
scan: scan-go scan-node scan-java

.PHONY: scan-go
scan-go: scan-go-govulncheck scan-go-nancy scan-go-osv-scanner

.PHONY: scan-go-govulncheck
scan-go-govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck -tags pkcs11 '$(go_dir)/...'

.PHONY: scan-go-nancy
scan-go-nancy:
	go install github.com/sonatype-nexus-community/nancy@latest
	go list -json -deps '$(go_dir)/...' | nancy sleuth

.PHONY: scan-go-osv-scanner
scan-go-osv-scanner:
	go install github.com/google/osv-scanner/cmd/osv-scanner@latest
	echo "GoVersionOverride = '$$(go env GOVERSION | sed 's/^go//')'" > osv-scanner.toml
	osv-scanner scan --lockfile='$(base_dir)/go.mod' || [ \( $$? -gt 1 \) -a \( $$? -lt 127 \) ]

.PHONY: scan-node
scan-node: scan-node-npm-audit scan-node-osv-scanner

.PHONY: scan-node-npm-audit
scan-node-npm-audit:
	cd '$(node_dir)' && \
		npm install --package-lock-only && \
		npm audit --omit=dev

.PHONY: scan-node-osv-scanner
scan-node-osv-scanner:
	go install github.com/google/osv-scanner/cmd/osv-scanner@latest
	cd '$(node_dir)' && \
		npm install --package-lock-only && \
		npm sbom --omit dev --package-lock-only --sbom-format cyclonedx > sbom.json && \
		osv-scanner scan --sbom=sbom.json

.PHONY: scan-java
scan-java: scan-java-dependency-check scan-java-osv-scanner

.PHONY: scan-java-dependency-check
scan-java-dependency-check:
	cd '$(java_dir)' && \
		mvn dependency-check:check -P owasp

.PHONY: scan-java-osv-scanner
scan-java-osv-scanner:
	go install github.com/google/osv-scanner/cmd/osv-scanner@latest
	osv-scanner scan --lockfile='$(java_dir)/pom.xml'

.PHONY: generate
generate:
	go install go.uber.org/mock/mockgen@latest
	go generate '$(go_dir)/...'

.PHONY: vendor-chaincode
vendor-chaincode:
	cd '$(scenario_dir)/fixtures/chaincode/golang/basic' && \
		GO111MODULE=on go mod vendor
	cd '$(scenario_dir)/fixtures/chaincode/golang/private' && \
		GO111MODULE=on go mod vendor

.PHONY: scenario-test-go
scenario-test-go: vendor-chaincode fabric-ca-client setup-softhsm
	cd '$(scenario_dir)/go' && \
		go test -timeout 20m -tags pkcs11 -v -args '$(scenario_dir)/features/'

.PHONY: scenario-test-go-no-hsm
scenario-test-go-no-hsm: vendor-chaincode
	cd '$(scenario_dir)/go' && \
		go test -timeout 20m -tags pkcs11 -v --godog.tags='~@hsm' -args '$(scenario_dir)/features/'

.PHONY: scenario-test-node
scenario-test-node: vendor-chaincode build-node fabric-ca-client setup-softhsm
	cd '$(scenario_dir)/node' && \
		rm -rf package-lock.json node_modules && \
		npm install && \
		npm test

.PHONY: scenario-test-node-no-hsm
scenario-test-node-no-hsm: vendor-chaincode build-node fabric-ca-client
	cd '$(scenario_dir)/node' && \
		rm -rf package-lock.json node_modules && \
		npm install && \
		npm run test:no-hsm

.PHONY: scenario-test-java
scenario-test-java: vendor-chaincode
	cd '$(java_dir)' && \
		mvn -Dmaven.javadoc.skip=true -DskipUnitTests verify

.PHONY: scenario-test
scenario-test: scenario-test-go scenario-test-node scenario-test-java

.PHONY: scenario-test-no-hsm
scenario-test-no-hsm: scenario-test-go-no-hsm scenario-test-node-no-hsm scenario-test-java

.PHONY: fabric-ca-client
fabric-ca-client:
	go install -tags pkcs11 github.com/hyperledger/fabric-ca/cmd/fabric-ca-client@latest

.PHONY: setup-softhsm
setup-softhsm:
	mkdir -p '$(TMPDIR)/softhsm'
	echo 'directories.tokendir = $(TMPDIR)/softhsm' > '$(SOFTHSM2_CONF)'
	softhsm2-util --init-token --slot 0 --label 'ForFabric' --pin 98765432 --so-pin 1234 || true

.PHONY: generate-docs
generate-docs:
	pip install --quiet --upgrade --requirement '$(base_dir)/requirements.txt'
	cd '$(base_dir)' && TZ=UTC mkdocs build --strict

.PHONY: generate-docs-node
generate-docs-node:
	cd '$(node_dir)' && \
		npm install && \
		npm run generate-apidoc

.PHONY: generate-docs-java
generate-docs-java:
	cd '$(java_dir)' && \
		mvn javadoc:javadoc

.PHONY: test
test: shellcheck unit-test scenario-test

.PHONY: all
all: test

.PHONY: clean
clean: clean-generated clean-node clean-java clean-docs

.PHONY: clean-node
clean-node:
	rm -rf '$(node_dir)/package-lock.json' '$(node_dir)/node_modules'

.PHONY: clean-java
clean-java:
	cd '$(java_dir)' && mvn clean

.PHONY: clean-generated
clean-generated:
	find '$(go_dir)' -name '*_mock_test.go' -delete

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
	cd '$(java_dir)' && mvn spotless:apply
