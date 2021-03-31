---
layout: home
---

The Fabric Gateway is a core component of a Fabric blockchain network and coordinates the actions required to
submit transactions and query ledger state on behalf of client applications.  By using the Gateway, client applications
only need to connect to a single endpoint in the Fabric network.

The Gateway SDKs implement the Fabric programming model as described in the
[Developing Applications](https://hyperledger-fabric.readthedocs.io/en/latest/developapps/developing_applications.html)
chapter of the Fabric documentation.

## Tech Preview

❇️ There are [samples for Go, Node, and Java](https://github.com/hyperledger/fabric-gateway/blob/main/samples/README.md) which are a great place to start if you want to try out the Fabric Gateway tech preview.

❇️ These include a Fabric network which includes a Peer with the Gateway feature enabled.

❇️ Make sure you [install the pre-reqs](#pre-reqs) before you begin.

## Overview

The original proposal is described in the [Fabric Gateway RFC](https://hyperledger.github.io/fabric-rfcs/text/0000-fabric-gateway.html).
Adding a Gateway component to the Fabric Peer provides a single entry point to a Fabric network, and removes much of the transaction submission logic from the client application.

The Gateway component in the Fabric Peer exposes a simple gRPC interface to client applications and manages the lifecycle of transaction invocation on behalf of the client.
This minimises the network traffic passing between the client and the blockchain network as well as minimising the number of network ports that need to be opened.

See the [gateway.proto file](https://github.com/hyperledger/fabric-protos/blob/main/gateway/gateway.proto) for details of the gRPC interface.

## Configuring the Gateway

Currently the Gateway component in the Fabric Peer is disabled by default. Enable the Gateway feature flag in `core.yaml` by adding the following:

```
peer:
    gateway:
        enabled: true
```

Alternatively, using [yq](https://mikefarah.gitbook.io/yq/):

```
docker run --rm -v "${PWD}":/workdir mikefarah/yq eval '.peer.gateway.enabled = true' --inplace core.yaml
```
## Client SDKs

Three SDKs are available to support the development of client applications that interact with the Fabric network via
the Gateway.  

### Go SDK

The Go SDK provides a high-level API for client applications written in Go.

See the following for more details:

- [Quickstart guide](https://github.com/hyperledger/fabric-gateway/blob/main/pkg/client/README.md) 
- [API documentation](https://pkg.go.dev/github.com/hyperledger/fabric-gateway/pkg/client)

### Node SDK

The Node SDK provides a high-level API for client applications written in Javascript or Typescript.

See the following for more details:

- [Quickstart guide](https://github.com/hyperledger/fabric-gateway/blob/main/node/README.md) 
- [API documentation](https://hyperledger.github.io/fabric-gateway/main/api/node/)

### Java SDK

The Java SDK provides a high-level API for client applications written in Java.

See the following for more details:

- [Quickstart guide](https://github.com/hyperledger/fabric-gateway/blob/main/java/README.md) 
- [API documentation](https://hyperledger.github.io/fabric-gateway/main/api/java/)

## Pre-reqs

Install the following pre-reqs to develop client applications using the Gateway SDK:

- Go v1.14 (required for sample Fabric network)
- Node (optional for Node SDK)
- Typescript (optional for Node SDK)
- Java 11 (optional for Java SDK)
- Docker (required for sample Fabric network)

In addition, you will need the `godog` tool to use the sample Fabric network, which can be installed with:

```
GO111MODULE=on go get github.com/cucumber/godog/cmd/godog@v0.10.0
```

Make sure you can run `godog --version` after installing. If the command is not found, add the Go bin directory to your path using: 

```
export PATH=$PATH:$(go env GOPATH)/bin
```
