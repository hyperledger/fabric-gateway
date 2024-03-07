# Hyperledger Fabric Gateway

For information on using the Fabric Gateway, including client API documentation, please visit the [Fabric Gateway documentation](https://hyperledger.github.io/fabric-gateway/).

For information on reporting issues, suggesting enhancements and contributing code, please review the [contributing guide](CONTRIBUTING.md).

## Overview

The original proposal is described in the [Fabric Gateway RFC](https://hyperledger.github.io/fabric-rfcs/text/0000-fabric-gateway.html).
Adding a gateway component to the Fabric peer provides a single entry point to a Fabric network, and removes much of the transaction submission logic from the client application.

The Gateway component in the Fabric Peer exposes a simple gRPC interface to client applications and manages the lifecycle of transaction invocation on behalf of the client.
This minimises the network traffic passing between the client and the blockchain network, as well as minimising the number of network ports that need to be opened.

See the [gateway.proto file](https://github.com/hyperledger/fabric-protos/blob/main/gateway/gateway.proto) for details of the gRPC interface.

## Building and testing

### Install pre-reqs

This repository comprises three functionally equivalent client APIs, written in Go, Typescript, and Java. In order to
build these components, the following need to be installed and available in the PATH:

- [Go 1.21+](https://go.dev/)
- [Node 18+](https://nodejs.org/)
- [Java 8+](https://adoptium.net/)
- [Docker](https://www.docker.com/)
- [Make](https://www.gnu.org/software/make/)
- [Maven](https://maven.apache.org/)
- [ShellCheck](https://github.com/koalaman/shellcheck#readme) (for linting shell scripts)

In order to run any of the Hardware Security Module (HSM) tests, [SoftHSM v2](https://www.opendnssec.org/softhsm/) is required. This can either be:

- installed using the package manager for your host system:
  - Ubuntu: `sudo apt install softhsm2`
  - macOS: `brew install softhsm`
  - Windows: **unsupported**
- or compiled and installed from source, following the [SoftHSM2 install instructions](https://wiki.opendnssec.org/display/SoftHSMDOCS/SoftHSM+Documentation+v2)
  - It is recommended to use the `--disable-gost` option unless you need **gost** algorithm support for the Russian market, since it requires additional libraries.

### Build using make

> **Note:** When the repository is first cloned, some mock implementations used for testing will not be present and the Go code will show compile errors. These will be generated when the `unit-test` target is run, or can be generated explicitly by running `make generate`.

The following Makefile targets are available:

- `make generate` - generate mock implementations used by unit tests
- `make lint` - run linting checks for the Go code
- `make unit-test-go` - run unit tests for the Go client API, excluding HSM tests
- `make unit-test-go-pkcs11` - run unit tests for the Go client API, including HSM tests
- `make unit-test-node` - run unit tests for the Node client API
- `make unit-test-java` - run unit tests for the Java client API
- `make unit-test` - run unit tests for all client language implementations
- `make pull-latest-peer` - fetch the latest peer docker image containing the gateway server
- `make scenario-test-go` - run the scenario (end to end integration) tests for Go client API, including HSM tests
- `make scenario-test-go-no-hsm` - run the scenario (end to end integration) tests for Go client API, excluding HSM tests
- `make scenario-test-node` - run the scenario tests for Node client API, including HSM tests
- `make scenario-test-node-no-hsm` - run the scenario tests for Node client API, excluding HSM tests
- `make scenario-test-java` - run the scenario tests for Java client API
- `make scenario-test` - run the scenario tests for all client language implementations
- `make scenario-test-no-hsm` - run the scenario tests for all client language implementations, excluding HSM tests
- `make shellcheck` - check for script errors
- `make test` - run all tests

### Scenario tests

The scenario tests create a Fabric network comprising two orgs (one peer in each org) and a single gateway within a set
of docker containers. The clients connect to the gateway to submit transactions and query the ledger state.

The tests are defined as feature files using the Cucumber BDD framework. The same set of feature files
is used across all three client language implementations to ensure consistency of behaviour.
