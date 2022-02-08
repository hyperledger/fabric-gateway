---
layout: home
---

The Fabric Gateway is a core component of a Hyperledger Fabric blockchain network, and coordinates the actions required to submit transactions and query ledger state on behalf of client applications. By using the Gateway, client applications only need to connect to a single endpoint in a Fabric network.

## Fabric Gateway v1.0

There are samples for Go, Node, and Java in the [fabric-samples](https://github.com/hyperledger/fabric-samples) repository, which are a great place to start if you want to try out the new Fabric Gateway!

- [fabric-samples/asset-transfer-basic](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-basic)
- [fabric-samples/asset-transfer-events](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-events)

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

The following table shows versions of Fabric, programming language runtimes, and other dependencies that are explicitly tested and that are supported for use with the Fabric Gateway client API.

|     | Tested | Supported |
| --- | ------ | --------- |
| **Fabric** | 2.4 | 2.4 |
| **Go** | 1.16 | 1.16 |
| **Node** | 14, 16 | 14, 16 |
| **Java** | 8, 11, 17 | 8, 11, 17 |
| **Platform** | Ubuntu 20.04 | |
