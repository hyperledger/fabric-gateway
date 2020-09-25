#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

base_dir := $(PWD)

go_dir := $(base_dir)/pkg
node_dir := $(base_dir)/node
scenario_dir := $(base_dir)/scenario

build: build-go build-node

build-go:
	go build -o bin/gateway prototype/gateway.go

build-node:
	cd $(node_dir); npm install

unit-test: unit-test-go unit-test-node

unit-test-go:
	go test -cover $(go_dir)/...

unit-test-node: build-node
	cd $(node_dir); npm test

lint:
	golint $(go_dir)/...

scenario-test-go: build-go
	cd $(scenario_dir)/go; godog $(scenario_dir)/features/

scenario-test-node: build
	cd $(scenario_dir)/node; npm install; npm test

scenario-test: scenario-test-go scenario-test-node

test: unit-test scenario-test

all: test
