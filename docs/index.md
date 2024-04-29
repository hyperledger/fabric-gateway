# Getting started

The Fabric Gateway is a core component of a Hyperledger Fabric blockchain network, and coordinates the actions required to submit transactions and query ledger state on behalf of client applications. By using the Gateway, client applications only need to connect to a single endpoint in a Fabric network. For a detailed description the Fabric Gateway, refer to the [architecture reference](https://hyperledger-fabric.readthedocs.io/en/latest/gateway.html) in the main Fabric documentation.

## Documentation

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

## Tutorials

The following tutorials describe how to write client applications using the Fabric Gateway client API:

- [Running a Fabric Application](https://hyperledger-fabric.readthedocs.io/en/latest/write_first_app.html) from the main Fabric documentation describes in detail the Fabric [asset-transfer-basic](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-basic#readme) sample.
- [Client Application Development](https://github.com/hyperledger/fabric-samples/tree/main/full-stack-asset-transfer-guide/docs/ApplicationDev#readme) section of the Fabric [full-stack-asset-transfer-guide](https://github.com/hyperledger/fabric-samples/tree/main/full-stack-asset-transfer-guide#readme) sample.

## Samples

There are samples for Go, Node, and Java in the [fabric-samples](https://github.com/hyperledger/fabric-samples) repository, which are a great place to start if you want to try out the Fabric Gateway.

- [asset-transfer-basic](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-basic#readme) for examples of transaction submit and evaluate.
- [asset-transfer-events](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-events#readme) for examples of chaincode eventing.
- [off_chain_data](https://github.com/hyperledger/fabric-samples/tree/main/off_chain_data#readme) for examples of block eventing.

## Migration

If migrating an existing application from one of the legacy Fabric client SDKs, consult the [migration guide](migration.md).
