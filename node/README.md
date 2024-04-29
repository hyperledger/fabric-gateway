# Hyperledger Fabric Gateway Client API for Node


The Fabric Gateway client API allows applications to interact with a Hyperledger Fabric blockchain network. It implements the Fabric programming model, providing a simple API to submit transactions to a ledger or query the contents of a ledger with minimal code.

## How to use

Samples showing how to create client applications that connect to and interact with a Hyperledger Fabric network, are available in the [fabric-samples](https://github.com/hyperledger/fabric-samples) repository:

- [asset-transfer-basic](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-basic) for examples of transaction submit and evaluate.
- [asset-transfer-events](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-events) for examples of chaincode eventing.
- [off_chain_data](https://github.com/hyperledger/fabric-samples/tree/main/off_chain_data) for examples of block eventing.

## API documentation

The Gateway client API documentation for Node is available here:

- https://hyperledger.github.io/fabric-gateway/main/api/node/

## Installation

Add a dependency to your project's `package.json` file with the command:

```sh
npm install @hyperledger/fabric-gateway
```

## Compatibility

This API requires Fabric v2.4 (or later) with a Gateway enabled Peer. Additional compatibility information is available in the documentation:

- https://hyperledger.github.io/fabric-gateway/
