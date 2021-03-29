# Fabric Gateway Samples

The samples in this repo show how to create client applications that invoke transactions using the new embedded Gateway
in Fabric.

The samples will only run against the latest Tech Preview version of Fabric.  The easiest way of setting up a gateway
enabled Fabric network is to use the scenario test framework that is part of this `fabric-gateway` repository using the
following command:

`make sample-network`

This will create a local docker network comprising five peers across three organisations and a single ordering node.
One of the peers (`peer0.org1.example.com`) has been configured with the gateway enabled.

A simple smart contract (named `basic`) will have been instantiated on all the peers.  The source code for the smart 
contract can examined [here](https://github.com/hyperledger/fabric-gateway/blob/main/scenario/fixtures/chaincode/golang/basic/main.go).

A sample client application is provided for each of the supported SDKs.
Note that the SDKs implement the Fabric 'Gateway' programming model which has been in use since
Fabric v1.4, but these are new implementations that target the embedded peer gateway and they share no common code with
existing Fabric SDKs.

In each of the language samples, the client application submits a transaction (`put`) to update the ledger followed by 
evaluating a transaction (`get`) to retrieve the value from the ledger (query).
The value that is being updated and retrieved is the current timestamp to demonstrate that the update is working.

### Go SDK

```
cd <base-path>/fabric-gateway/samples/go
go run sample.go
```

### Node SDK

```
cd <base-path>/fabric-gateway/samples/node
npm install
npm run build
npm start
```

### Java

TBD