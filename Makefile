#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

base_dir := $(PWD)
node_sdk_dir := $(base_dir)/client/node/sdk
go_sdk_dir := $(base_dir)/client/go/sdk

build:
	go build -o bin/gateway prototype/gateway.go

unit-test:
	go test -cover ./...

lint:
	golint ./...

test-scenario-sdk-go: build
	cd $(go_sdk_dir)/scenario; godog $(base_dir)/scenario/features/

test-scenario-sdk-node: build
	cd $(node_sdk_dir); npm install; ./node_modules/.bin/cucumber-js --require './steps/**/*.js' $(base_dir)/scenario/features/*.feature

test-scenario: test-scenario-sdk-go test-scenario-sdk-node

test: unit-test test-scenario

all: test
