# Hyperledger Fabric Gateway

[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/7278/badge)](https://www.bestpractices.dev/projects/7278)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/hyperledger/fabric-gateway/badge)](https://scorecard.dev/viewer/?uri=github.com/hyperledger/fabric-gateway)
[![Go Report Card](https://goreportcard.com/badge/github.com/hyperledger/fabric-gateway)](https://goreportcard.com/report/github.com/hyperledger/fabric-gateway)

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

- [Go 1.23+](https://go.dev/)
- [Node 18+](https://nodejs.org/)
- [Java 11+](https://adoptium.net/)
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

#### Dev Container

This project includes a [Dev Container](https://containers.dev/) configuration that includes all of the pre-requisite software described above in a Docker container, avoiding the need to install them locally. The only requirement is that [Docker](https://www.docker.com/) is installed and available.

Opening the project folder in an IDE such as [VS Code](https://code.visualstudio.com/docs/devcontainers/containers) (with the [Dev Containers extention](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)) should offer the option of opening in the Dev Container. Alternatively, VS Code allows the remote repository to [opened directly in an isolated Dev Container](https://code.visualstudio.com/docs/devcontainers/containers#_quick-start-open-a-git-repository-or-github-pr-in-an-isolated-container-volume).

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
- `make scenario-test-go` - run the scenario (end to end integration) tests for Go client API, including HSM tests
- `make scenario-test-go-no-hsm` - run the scenario (end to end integration) tests for Go client API, excluding HSM tests
- `make scenario-test-node` - run the scenario tests for Node client API, including HSM tests
- `make scenario-test-node-no-hsm` - run the scenario tests for Node client API, excluding HSM tests
- `make scenario-test-java` - run the scenario tests for Java client API
- `make scenario-test` - run the scenario tests for all client language implementations
- `make scenario-test-no-hsm` - run the scenario tests for all client language implementations, excluding HSM tests
- `make shellcheck` - check for script errors
- `make test` - run all tests
- `make format` - fix all code formatting

### Scenario tests

The scenario tests create a Fabric network comprising two orgs (one peer in each org) and a single gateway within a set
of docker containers. The clients connect to the gateway to submit transactions and query the ledger state.

The tests are defined as feature files using the Cucumber BDD framework. The same set of feature files
is used across all three client language implementations to ensure consistency of behaviour.

### Documentation

The documentation site is built using [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/). Documentation build configuration is in [mkdocs.yml](mkdocs.yml), and the site content is in the [docs](docs) folder.
