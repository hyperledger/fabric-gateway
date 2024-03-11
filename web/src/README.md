# Overview

The _fabric-gateway-web_ bundle enables Web developers to create signed transaction artifacts in in browser applications. These artifacts can then be serialized and sent to an intermediary service to interact with Fabric on behalf of the browser client, using the [Fabric Gateway client API](https://hyperledger.github.io/fabric-gateway/).

## Getting started

A session for a given client identity is created by calling `connect()` with a client identity, and client signing implementation. The returned `Gateway` enables interaction with any of the blockchain `Networks` (channels) accessible through the Fabric Gateway. This in turn provides access to Smart `Contracts` within chaincode deployed to that blockchain network, and to which transactions can be submitted or queries can be evaluated.

To **evaluate** a smart contract transaction function, querying ledger state:

1. The client:
   1. Creates a signed transaction `Proposal`.
   1. Serializes the `Proposal` and sends it to the intermediary service.
1. The intermediary service:
   1. Deserializes the data into a Proposal object.
   1. On behalf of the client, _evaluates_ the transaction proposal.
   1. Returns the response to the client.

To **submit** a transaction, updating ledger state:

1. The client:
   1. Creates a signed transaction `Proposal`.
   1. Serializes the `Proposal` and sends it to the intermediary service.
1. The intermediary service:
   1. Deserializes the data into a Proposal object.
   1. On behalf of the client, _endorses_ the transaction proposal.
   1. Serializes the resulting Transaction object and returns it to the client.
1. The client:
   1. Deserializes the data to create a signed `Transaction`.
   1. Serializes the signed `Transaction` and sends it to the intermediary service.
1. The intermediary service:
   1. Deserializes the data into a signed Transaction object.
   1. On behalf of the client, _submits_ the transaction.
   1. Waits for the transaction commit status.
   1. Returns the result to the client.

## Example

A Gateway connection is created using a client identity and a signing implementation:

```JavaScript
const identity = {
    mspId: 'myorg',
    credentials,
};

const signer = async (message) => {
    const signature = await globalThis.crypto.subtle.sign(
        { name: 'ECDSA', hash: 'SHA-256' },
        privateKey,
        message,
    );
    return new Uint8Array(signature);
};

const gateway = connect({ identity, signer });
```

The following example shows how to create a signed transaction proposal. The serialized proposal can be sent to an intermediary service, which can then use it to interact with Fabric on the Web client's behalf.

```JavaScript
const network = gateway.getNetwork('channelName');
const contract = network.getContract('chaincodeName');

const proposal = await contract.newProposal('transactionName', {
    arguments: ['one', 'two'],
});
const proposalBytes = proposal.getBytes();
```

The following example shows how to create a signed transaction from an endorsed transaction received from an intermediary service following proposal endorsement. The transaction results can be inspected, and the serialized transaction can be sent to an intermediary service, which can then submit it to Fabric to update the ledger.

```JavaScript
const transaction = await gateway.newTransaction(endorsedTransactionBytes);
if (transaction.getTransactionId() !== proposal.getTransactionId()) {
    // Not the expected response so might be from a malicious actor.
}

const result = transaction.getResult();
const transactionBytes = transaction.getBytes();
```
