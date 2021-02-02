# Hyperledger Fabric Gateway

The Fabric Gateway is a core component of a Fabric blockchain network and coordinates the actions required to 
submit transactions and query ledger state on behalf of client applications.  By using the Gateway, client applications
only need to connect to a single endpoint in the Fabric network.

The Gateway SDKs implement the Fabric programming model as described in the 
[Developing Applications](https://hyperledger-fabric.readthedocs.io/en/latest/developapps/developing_applications.html)
chapter of the Fabric documentation.

## The Gateway server

The Gateway server is a standalone process (written in Go) and exposes a simple gRPC interface to client applications.
The server manages the lifecycle of transaction invocation on behalf of the client, minimising the network traffic passing
between the client and the blockchain network as well as minimising the number of network ports that need to be opened.

A prebuilt docker image is available for the Gateway server.

## Client SDKs

Three SDKs are available to support the development of client applications that interact with the Fabric network via
the Gateway.  

### Go SDK

The Go SDK provides a high-level API for client applications written in Go.

Read the [quickstart](pkg/client/README.md) guide for more details.

### Node SDK

The Node SDK provides a high-level API for client applications written in Javascript or Typescript.

Read the [quickstart](node/README.md) guide for more details.

### Java SDK

The Java SDK provides a high-level API for client applications written in Java.

Read the [quickstart](java/README.md) guide for more details.


## Building and testing

#### Install pre-reqs

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

#### Build using make

The following Makefile targets are available
- `make build-go` - compile the gateway server executable
- `make docker` - create a docker image containing the gateway server
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

#### Scenario tests

The scenario tests create a Fabric network comprising two orgs (one peer in each org) and a single gateway within a set
of docker containers.  The clients connect to the gateway to submit transactions and query the ledger state.

The tests are defined as feature files using the Cucumber BDD framework.  The same set of feature files
is used across all three SDKs to ensure consistency of behaviour. 
