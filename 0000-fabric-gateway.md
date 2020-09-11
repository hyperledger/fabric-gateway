---
layout: default
title: Fabric Gateway
nav_order: 3
---

- Feature Name: Fabric Gateway
- Start Date: 2020-09-08
- RFC PR: (leave this empty)
- Fabric Component: fabric-gateway
- Fabric Issue: (leave this empty)

# Summary
[summary]: #summary

The Fabric Gateway is a new component that will run either as a standalone process, or embedded with the peer, and will implement as much of the high-level 'gateway' programming model as possible.  This will remove much of the transaction submission logic from the client application which will simplify the maintenance of the SDKs.  It will also simplify the administrative overhead of running a Fabric network since the client application will only need to connect to a gateway rather than multiple peers across potentially multiple organisations.

# Motivation
[motivation]: #motivation

The high-level "gateway" programming model has proved to be very popular with users since its introduction in v1.4 in Node SDK.  Since then, it has been implemented in the Java and Go SDKs.  This has led to a large amount of code that has to be maintained in the SDKs which increases the cost of implementing new feature which have to be replicated to keep the SDKs in step.  It also increases the cost of creating an SDK for a new language.  Moving all of the high-level logic into its own process and exposing a simplified GRPC interface will drastically reduce the complexity of the SDKs.

By providing a single entry point to a Fabric network, client applications can interact with complex network topologies without multiple ports having to be opened and secured to each peer and ordering node.  The client will connect only to the gateway which will act on its behalf with the rest of the network.

# Guide-level explanation
[guide-level-explanation]: #guide-level-explanation

The Fabric Gateway is an embodiment of the high-level 'gateway' Fabric programming model in a server component that will form part of a Fabric network alongside Peers, Orderers and CAs.  It will either be stood up inside its own process in a docker container, or it will be hosted inside a peer.  Either way, it exposes its functionality to clients via a GRPC interface.

Lightweight client SDKs are used to connect to the Gateway for the purpose of invoking transactions.  The gateway will intereact with the rest of the network on behalf of the client eliminating the need for the client application to connect directly to peers and orderers.

The concepts and APIs will be familiar to Fabric programmers who are already using the new programming model.


# Reference-level explanation
[reference-level-explanation]: #reference-level-explanation

The Gateway runs as a client process associated with an organisation and requires an identity (issued by the org's CA) in order to interact with the discovery service and to register for events (Deliver Client).  

The Gateway also runs as a GRPC server and exposes the following services to client applications (SDKs):

```
service Gateway {
    rpc Prepare(SignedProposal) returns (PreparedTransaction) {}
    rpc Commit(PreparedTransaction) returns (stream Event) {}
    rpc Evaluate(SignedProposal) returns (Result) {}
}
```

Submitting a transaction is a two step process (performed by the client SDK):
- The client creates a SignedProposal message as defined in `fabric-protos/peer/proposal.proto` and signed with the user's identity
- The `Prepare` service is invoked on the Gateway, passing the SignedProposal message
  - The Gateway will determine the endorsement plan for the requested chaincode and forward to the appropriate peers for endorsement. It will return to the client a `PreparedTransaction` message which contains a `Envelope` message as defined in `fabric-protos/common/common.proto`.  It will also contain other information, such as the return value of the chaincode function and the transaction ID so that the client doesn't necessarily need to unpack the `Envelope` payload.
- The client signs the hash of the Envelope payload using the user's private key and sets the signature field in the `Envelope` in the `PreparedTransaction`.
- The `Commit` service is invoked on the Gateway, passing the signed `PreparedTransaction` message.  A stream is opened to return multiple return values.
  - The Gateway will register transaction event listeners for the given channel/txId.
  - It will then broadcast the `Envelope` to the ordering service.
  - The success/error response is passed back to the client in the stream
  - The Gateway awaits suffient tx commit events before returning and closing the stream, indicating to the client that transaction has been committed.

Evaluating a transaction is a simple process of invoking the `Evaluate` service passing a SignedProposal message.  The Gateway passes the request to a peer of it's choosing according to a defined policy (probably same org, highest block count) and returns the chaincode function return value to the client.

## Launching the Gateway

To run the gateway server, the following parameters need to be supplied:
- The url of at least one peer in the org.  Once connected, the discovery service will be invoked to find other peers.
- The MSPID associated with the gateway.
- The signing identity of the gateway (cert and key), e.g. location of PEM files on disk.
- The TLS certificate so that the gateway can connect to other components in the organization
The following is an example command line invocation using a prototype:
- `gateway -h peer0.org1.example.com -p 7051 -m Org1MSP -id ../../fabric-samples/fabcar/javascript/wallet/gateway.id -tlscert ../../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem`

Alternatively, the Gateway server will be designed to be embeddable within a Peer, in which case the all of the above parameters will be taken from the peer configuration.

## Scaling and load balancing

Currently a client application is responsible for load balancing its requests between multiple peers.  The Gateway will handle this on behalf of the client application.  The Gateway will be able to maintain information on block height of each peer in order to optimise the routing of queries.

The Gateway itself will be stateless (i.e. not maintain client session state) and so can be scaled using an appriopriate load balancer.

## Gateway SDKs

A set of SDKs will be created to allow client applications to interact with the Fabric Gateway.  The APIs exposed by these SDKs will be, where possible, the same as the current high-level 'gateway' SDKs.

The following classes/structures will be available as part of the gateway programming model:
- Gateway
- Network
- Contract
- Transaction
- Wallet

The existing `submitTransaction()` and `evaluateTransaction()` API methods will be implemented using the Gateway GRPC services described above.

Based on user feedback of the existing gateway SDKs, the following additions are planned to be supported by this implementation:
- Support for 'offline signing', allowing clients to write REST servers that don't directly have access to users' private keys.
- Support for 'asynchronous submit', allowing client applications to continue processing while the transaction commit proceeds in the background. For Go this will be implemented by returning a channel; for Node by returning a Promise; and for Java by returning a Future.

SDKs will be created for the following languages:
- Node (Typescript/Javascript)
- Go
- Java

It is proposed that these SDKs will be maintained in the same GitHub respository as the gateway itself (`hyperledger/fabric-gateway`) to ensure that they all stay up to date with each other and with the core gateway.  This will be enforced in the CI pipeline by running the end-to-end scenario tests across all SDKs.

Contributors are invited to create SDKs for other languages, e.g. Python, Rust.


# Drawbacks
[drawbacks]: #drawbacks

When run as a standalone process, the gateway would add another docker containter to an (arguably) already complex Fabric architecture.  This could be mitigated by embedding the gateway server within the organisation's existing peers.  This feature will be designed to be embeddable, if required.

# Rationale and alternatives
[alternatives]: #alternatives

- Why is this design the best in the space of possible designs?
- What other designs have been considered and what is the rationale for not
  choosing them?
- What is the impact of not doing this?

# Prior art
[prior-art]: #prior-art

The Gateway SDK has been available as the high level programming model in the Fabric SDKs since v1.4.  It has been thoroughly tested in manu scenarios.  This Fabric Gateway component is an embodiment of that programming model in a server component.

# Testing
[testing]: #testing

In addition to the usual unit tests for all functions, a comprehensive end to end 
scenario test suite will be created from the outset.  A set of 'cucumber' BDD style 
tests will be taken from the existing Java and Node SDKs to test all the major functions
of the gateway against a multi-org multi-peer network.  These tests will include:
- evaluation and submission of transactions
- event handling scenarios
- discovery
- transient / private data
- error handling

# Dependencies
[dependencies]: #dependencies



# Unresolved questions
[unresolved]: #unresolved-questions

- What parts of the design do you expect to resolve through the RFC process
  before this gets merged?
- What parts of the design do you expect to resolve through the implementation
  of this feature before stabilization?
- What related issues do you consider out of scope for this RFC that could be
  addressed in the future independently of the solution that comes out of this
  RFC?
