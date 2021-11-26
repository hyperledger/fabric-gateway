---
layout: home
---

The Fabric Gateway is a core component of a Fabric blockchain network and coordinates the actions required to submit transactions and query ledger state on behalf of client applications.
By using the Gateway, client applications only need to connect to a single endpoint in the Fabric network.

The Gateway SDKs implement the Fabric programming model as described in the [Developing Applications](https://hyperledger-fabric.readthedocs.io/en/latest/developapps/developing_applications.html) chapter of the Fabric documentation.

## Fabric Gateway v1.0

❇️ There are [samples for Go, Node, and Java](https://github.com/hyperledger/fabric-gateway/blob/main/samples/README.md) which are a great place to start if you want to try out the new Fabric Gateway!

❇️ Make sure you [install the pre-reqs](#pre-reqs) before you begin.

## Client SDKs

Three SDKs are available to support the development of client applications that interact with the Fabric network via the Gateway.  

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

- Go v1.16.7 (required sample Fabric network and Go SDK)
- Node 14.x (required for Node SDK)
- Typescript (required for Node SDK)
- Java 8 (required for Java SDK)
- Docker (required for sample Fabric network)

In addition, you will need the `godog` tool to use the sample Fabric network, which can be installed with:

```
GO111MODULE=on go get github.com/cucumber/godog/cmd/godog@v0.10.0
```

Make sure you can run `godog --version` after installing. If the command is not found, add the Go bin directory to your path using: 

```
export PATH=$PATH:$(go env GOPATH)/bin
```
