# Hyperledger Fabric Gateway

For information on using the Fabric Gateway, including client API documentation, please visit the [Fabric Gateway documentation](https://hyperledger.github.io/fabric-gateway/).

## Overview

The original proposal is described in the [Fabric Gateway RFC](https://hyperledger.github.io/fabric-rfcs/text/0000-fabric-gateway.html).
Adding a gateway component to the Fabric peer provides a single entry point to a Fabric network, and removes much of the transaction submission logic from the client application.

The Gateway component in the Fabric Peer exposes a simple gRPC interface to client applications and manages the lifecycle of transaction invocation on behalf of the client.
This minimises the network traffic passing between the client and the blockchain network, as well as minimising the number of network ports that need to be opened.

See the [gateway.proto file](https://github.com/hyperledger/fabric-protos/blob/main/gateway/gateway.proto) for details of the gRPC interface.

## Building and testing

### Install pre-reqs

This repository comprises three functionally equivalent client APIs, written in Go, Typescript, and Java. In order to
build these components, the following needs to be installed and available in the PATH:
- Go 1.17
- Node 14
- Java 8
- Docker
- Make

Additional required tools are installed using the following Makefile target:

- `make setup`

In order to run any of the Hardware Security Module (HSM) tests, the following must also be installed:

- pkcs11 enabled fabric-ca-client
  - `go get -tags 'pkcs11' github.com/hyperledger/fabric-ca/cmd/fabric-ca-client`
- SoftHSM, which can either be:
  - installed using the package manager for your host system:
    - Ubuntu: `sudo apt install softhsm2`
    - macOS: `brew install softhsm`
    - Windows: **unsupported**
  - or compiled and installed from source:
    1. install openssl 1.0.0+ or botan 1.10.0+
    2. download the source code from <https://dist.opendnssec.org/source/softhsm-2.5.0.tar.gz>
    3. `tar -xvf softhsm-2.5.0.tar.gz`
    4. `cd softhsm-2.5.0`
    5. `./configure --disable-gost` (would require additional libraries, turn it off unless you need 'gost' algorithm support for the Russian market)
    6. `make`
    7. `sudo make install`

### Build using make

> **Note:** When the repository is first cloned, some mock implementations used for testing will not be present and the Go code will show compile errors. These will be generated when the `unit-test` target is run, or can be generated explicitly by running `make generate`.

The following Makefile targets are available:
- `make generate` - generate mock implementations used by unit tests
- `make unit-test-go` - run unit tests for the Go client API
- `make unit-test-node` - run unit tests for the Node client API
- `make unit-test-java` - run unit tests for the Java client API
- `make unit-test` - run unit tests for all client language implementations
- `make pull-latest-peer` - fetch the latest peer docker image containing the gateway server
- `make scenario-test-go` - run the scenario (end to end integration) tests for Go client API
- `make scenario-test-node` - run the scenario tests for Node client API
- `make scenario-test-java` - run the scenario tests for Java client API
- `make scenario-test` - run the scenario tests for all client language implementations
- `make test` - run all unit and scenario tests

### Scenario tests

The scenario tests create a Fabric network comprising two orgs (one peer in each org) and a single gateway within a set
of docker containers.  The clients connect to the gateway to submit transactions and query the ledger state.

The tests are defined as feature files using the Cucumber BDD framework.  The same set of feature files
is used across all three client language implementations to ensure consistency of behaviour.

### Run Samples

Refer to [Fabric-Samples](https://github.com/hyperledger/fabric-samples) for sample applications developed using fabic-gateway.
