# Hyperledger Fabric Gateway

For information on using the Gateway, including client SDK API references, please visit the [Fabric Gateway documentation](https://hyperledger.github.io/fabric-gateway/).

## Overview

The original proposal is described in the [Fabric Gateway RFC](https://hyperledger.github.io/fabric-rfcs/text/0000-fabric-gateway.html).
Adding a Gateway component to the Fabric Peer provides a single entry point to a Fabric network, and removes much of the transaction submission logic from the client application.

The Gateway component in the Fabric Peer exposes a simple gRPC interface to client applications and manages the lifecycle of transaction invocation on behalf of the client.
This minimises the network traffic passing between the client and the blockchain network as well as minimising the number of network ports that need to be opened.

See the [gateway.proto file](https://github.com/hyperledger/fabric-protos/blob/main/gateway/gateway.proto) for details of the gRPC interface.

## Building and testing

### Install pre-reqs

This repo comprises the Gateway server (written in Go) and three SDKs (written in Go, Typescript and Java).
In order to build these components, the following needs to be installed and available in the PATH:
- Go (v1.14)
- Node (optional for Node SDK)
- Typescript (optional for Node SDK)
- Java 11 (optional for Java SDK)
- Docker
- Protobuf compiler (https://developers.google.com/protocol-buffers/docs/downloads)
- Some Go tools:
  - `GO111MODULE=on go get github.com/cucumber/godog/cmd/godog@v0.10.0`
  - `go get -u golang.org/x/lint/golint`
  - `go get -u golang.org/x/tools/cmd/goimports`
  - `go get google.golang.org/grpc google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc`
  - `go get honnef.co/go/tools/cmd/staticcheck`
  - `go get github.com/golang/mock/mockgen`
- pkcs11 enabled fabric-ca-client
  - `go get -tags 'pkcs11' github.com/hyperledger/fabric-ca/cmd/fabric-ca-client`
- SoftHSM which can be installed using the package manager for your host system:
* Ubuntu: `sudo apt install softhsm2`
* macOS: `brew install softhsm`
* Windows: **unsupported**

Or compiled and installed from source:

1. install openssl 1.0.0+ or botan 1.10.0+
2. download the source code from <https://dist.opendnssec.org/source/softhsm-2.5.0.tar.gz>
3. `tar -xvf softhsm-2.5.0.tar.gz`
4. `cd softhsm-2.5.0`
5. `./configure --disable-gost` (would require additional libraries, turn it off unless you need 'gost' algorithm support for the Russian market)
6. `make`
7. `sudo make install`

### Build using make

The following Makefile targets are available
- `make build-go` - compile the gateway server executable
- `make pull-latest-peer` - fetch the latest peer docker image containing the gateway server
- `make unit-test-go` - run unit tests for the gateway server and Go SDK
- `make unit-test-node` - run unit tests for the Node SDK
- `make unit-test-java` - run unit tests for the Java SDK
- `make unit-test` - run unit tests for the gateway server and all three SDKs
- `make scenario-test-go` - run the scenario (end to end integration) tests for Go SDK
- `make scenario-test-node` - run the scenario tests for Node SDK
- `make scenario-test-java` - run the scenario tests for Java SDK
- `make scenario-test` - run the scenario tests for all SDKs
- `make test` - run all unit and scenario tests
- `make generate` - generate mock implementations used by unit tests

Note that immediately after creating a fresh copy of this repository, auto-generated test mocks will not be preset so
Go code will show errors. Running the `unit-test` make target will generate the required mock implementations, and they
can also be generated explicitly by running `make generate`.

### Scenario tests

The scenario tests create a Fabric network comprising two orgs (one peer in each org) and a single gateway within a set
of docker containers.  The clients connect to the gateway to submit transactions and query the ledger state.

The tests are defined as feature files using the Cucumber BDD framework.  The same set of feature files
is used across all three SDKs to ensure consistency of behaviour.
