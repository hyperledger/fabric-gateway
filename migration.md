---
title: "Migration guide"
---

This page documents key considerations when rewriting an existing application, written using legacy Hyperledger Fabric client SDKs, to the Fabric Gateway client API.

## Fabric programming model

The Fabric Gateway client API is an evolution of the legacy SDKs and the Fabric programming model. The API structure and capability remain broadly the same as the legacy SDKs. Similarities include:

- **Gateway**: connection to Fabric peer(s) providing access to blockchain networks.
- **Network**: blockchain network of nodes hosting a shared ledger (analogous to a channel).
- **Contract**: smart contract deployed to a blockchain network.
- **Submit transaction**: invoke a smart contract transaction function to update ledger state.
- **Evaluate transaction**: invoke a smart contract transaction function to query ledger state.
- **Chaincode events**: receive events emitted by committed transactions to trigger business processes.
- **Block events**: receive blocks committed to the ledger.
- **Event checkpointing**: persist current event position to support resume of eventing.

The high level API to connect a Gateway instance, and submit or evaluate a transaction remains almost identical.

For more advanced transaction invocations, such as those involving transient data, the legacy SDKs provide a `createTransaction()` method on the Contract object, which allows the client application to specify additional invocation parameters (see [Go](https://pkg.go.dev/github.com/hyperledger/fabric-sdk-go/pkg/gateway?utm_source=godoc#Contract.CreateTransaction), [Node](https://hyperledger.github.io/fabric-sdk-node/release-2.2/module-fabric-network.Contract.html#createTransaction), and [Java](https://hyperledger.github.io/fabric-gateway-java/release-2.2/org/hyperledger/fabric/gateway/Contract.html#createTransaction(java.lang.String)) documentation). The Fabric Gateway client API provides a `newProposal()` method on the Contract object to perform the same function (see [Go](https://pkg.go.dev/github.com/hyperledger/fabric-gateway/pkg/client#Contract.NewProposal), [Node](https://hyperledger.github.io/fabric-gateway/main/api/node/interfaces/Contract.html#newProposal), and [Java](https://hyperledger.github.io/fabric-gateway/main/api/java/org/hyperledger/fabric/client/Contract.html#newProposal(java.lang.String)) documentation).

## Key differences

The key API and behavioral differences that need to be considered when switching from legacy SDKs to the Fabric Gateway client API are:

- **[gRPC connections](#grpc-connections)** are managed by the application, and can be shared by Gateway instances.
- **[Connection profiles](#connection-profiles)** are not needed.
- **[Wallets](#wallets)** are not needed, with the application choosing how to manage credential storage.
- **[Endorsement requirements](#endorsement-requirements)** generally no longer need to be specified.
- **[Event reconnect](#event-reconnect)** is controlled by the client application.

More detail and recommendations for each of these items is provided below.

### gRPC connections

In the legacy SDKs, each Gateway instance maintains internal gRPC connections to network nodes used to evaluate and submit transactions, and to obtain events. Many gRPC connections may be created for each Gateway instance, and these connections are not shared with other Gateway instances. Since there is significant overhead associated with creating gRPC connections, this can cause resource issues.

In the Fabric Gateway client API, each Gateway instance uses a single gRPC connection to the Fabric Gateway service for all operations. The Gateway instance's gRPC connection is provided by the client application, and it can be shared by multiple Gateway instances. This allows the client application complete control of gRPC connection configuration and resource allocation.

The API documentation contains examples of creating a gRPC connection and using this to connect a Gateway instance for [Go](https://pkg.go.dev/github.com/hyperledger/fabric-gateway/pkg/client#example-package), [Node](https://hyperledger.github.io/fabric-gateway/main/api/node/#example) and [Java](https://hyperledger.github.io/fabric-gateway/main/api/java/).

### Connection profiles

The Fabric Gateway client API does not use common connection profiles. Instead, only the endpoint address of the Fabric Gateway service is required to establish a gRPC connection that will be used when connecting a Gateway instance. Since the Fabric Gateway service is provided by Fabric peers, the endpoint address may be one of the peer addresses that would be defined in a connection profile. It could also be the address of a load balancer or ingress controller that forwards connections to network peers, providing high availability.

### Wallets

The legacy SDKs provide wallets for credential management. Wallets perform two functions:

1. Persistent credential storage.
1. Configuration of the Gateway client based on the type of credentials (for example, identities managed by a Hardware Security Module).

Using the Fabric Gateway client API, the mechanism for storing credentials is a choice for the client application. The application may continue using the legacy SDKs to access credentials stored in a wallet, or may use a different mechanism for storing and accessing credentials.

To connect a Gateway instance, the application simply provides an Identity object and a signing implementation. Helper functions are provided to create an Identity object from an X.509 certificate, and also to create a signing implementation from either a private key or an HSM-managed identity. To make use of alternative signing mechanisms, the application may provide its own signing implementation.

### Endorsement requirements

When using the legacy SDKs in more complex scenarios, such as those involving private data collections, chaincode-to-chaincode calls, or key-based endorsement policies, it is often necessary for the client application to explicitly specify endorsement requirements for a transaction invocation. This may be in the form of specifying chaincode interests, endorsing organizations, or endorsing peers.

Using the Fabric Gateway client API, it is generally not necessary for the client application to specify endorsement requirements. The Fabric Gateway service dynamically determines the endorsement requirements for a given transaction invocation and uses the most appropriate peers to obtain endorsement.

For transaction proposals that contain transient data, there are two notable scenarios that do require the application to explicitly specify the organizations that may be used for endorsement:

1. The Fabric Gateway service's organization is unable to endorse the transaction proposal.
1. Transactions that perform blind writes to private data collections from which they do not have read permission.

These restrictions are in place to ensure that private data is not distributed to organizations that should not have access to it.

It is recommended to only specify endorsing organizations in cases where it is specifically required.

### Event reconnect

In the event of a peer or network failure during event listening, the legacy SDKs transparently attempt to reestablish connection and continue delivering events once successful reconnection is achieved. The Fabric Gateway client API surfaces eventing errors to the client application at the point it requests the next event. To reestablish eventing, the application must initiate a new event listening session with an appropriate start position.

Event checkpointing tracks the current event position and can be used to resume eventing at the correct start position on reconnect.
