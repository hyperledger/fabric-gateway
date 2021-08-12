# Hyperledger Fabric Gateway Client SDK for Go


The Fabric Gateway SDK allows applications to interact with a Fabric blockchain network.  It provides a simple API to submit transactions to a ledger or query the contents of a ledger with minimal code.

The Gateway SDK implements the Fabric programming model as described in the [Developing Applications](https://hyperledger-fabric.readthedocs.io/en/latest/developapps/developing_applications.html) chapter of the Fabric documentation.

## How to use

Samples showing how to create a client application that updates and queries the ledger
are available for each of the supported SDK languages here:
https://github.com/hyperledger/fabric-gateway/tree/main/samples

### API documentation

The Go Gateway SDK documentation is available here:
https://pkg.go.dev/github.com/hyperledger/fabric-gateway/pkg/client

### Installation

`go get github.com/hyperledger/fabric-gateway`

### Compatibility

This SDK requires Fabric 2.4 with a Gateway enabled Peer.
