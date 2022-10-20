---
layout: home
---

The Fabric Gateway is a core component of a Hyperledger Fabric blockchain network, and coordinates the actions required to submit transactions and query ledger state on behalf of client applications. By using the Gateway, client applications only need to connect to a single endpoint in a Fabric network.

## Fabric Gateway v1.1

There are samples for Go, Node, and Java in the [fabric-samples](https://github.com/hyperledger/fabric-samples) repository, which are a great place to start if you want to try out the new Fabric Gateway!

- [fabric-samples/asset-transfer-basic](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-basic) for examples of transaction submit and evaluate.
- [fabric-samples/asset-transfer-events](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-events) for examples of chaincode eventing.
- [fabric-samples/off_chain_data](https://github.com/hyperledger/fabric-samples/tree/main/off_chain_data) for examples of block eventing.

If migrating an existing application from one of the legacy Fabric client SDKs, consult the [migration guide](migration).

## Client API

The Fabric Gateway client API is available for several programming languages to support the development of client applications that interact with a Fabric network using the Gateway.  

### Go

- [Quickstart guide](https://github.com/hyperledger/fabric-gateway/blob/main/pkg/client/README.md) 
- [API documentation](https://pkg.go.dev/github.com/hyperledger/fabric-gateway/pkg/client)

### Node

- [Quickstart guide](https://github.com/hyperledger/fabric-gateway/blob/main/node/README.md) 
- [API documentation](https://hyperledger.github.io/fabric-gateway/main/api/node/)

### Java

- [Quickstart guide](https://github.com/hyperledger/fabric-gateway/blob/main/java/README.md) 
- [API documentation](https://hyperledger.github.io/fabric-gateway/main/api/java/)

## Compatibility

Each minor release version of Fabric Gateway client API targets the current supported versions of Go, and the current long-term support (LTS) releases of Node and Java. A specific minimum version of Hyperledger Fabric for the Gateway peer is also required for full functionality.

The following table shows versions of Fabric, programming language runtimes, and other dependencies that are explicitly tested and that are supported for use with the Fabric Gateway client API.

|     | Tested | Supported |
| --- | ------ | --------- |
| **Fabric** | 2.4 | 2.4.4+ |
| **Go** | 1.17, 1.18, 1.19 | 1.17, 1.18, 1.19 |
| **Node** | 14, 16, 18 | 14, 16, 18 |
| **Java** | 8, 11, 17 | 8, 11, 17 |
| **Platform** | Ubuntu 22.04 | |
