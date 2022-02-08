# Hyperledger Fabric Gateway Client API for Go

The Fabric Gateway client API allows applications to interact with a Hyperledger Fabric blockchain network. It implements the Fabric programming model, providing a simple API to submit transactions to a ledger or query the contents of a ledger with minimal code.

## How to use

Samples showing how to create a client application that updates and queries the ledger, and listens for events, are available in the [fabric-samples](https://github.com/hyperledger/fabric-samples) repository:

* [fabric-samples/asset-transfer-basic](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-basic)
* [fabric-samples/asset-transfer-events](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-events)

## API documentation

The Gateway client API documentation for Go is available here:

* https://pkg.go.dev/github.com/hyperledger/fabric-gateway/pkg/client

## Installation

Add a package dependency to your project with the command:

```sh
go get github.com/hyperledger/fabric-gateway
```

## Compatibility

This API requires Fabric 2.4 with a Gateway enabled Peer. Additional compatibility information is available in the documentation:

* https://hyperledger.github.io/fabric-gateway/
