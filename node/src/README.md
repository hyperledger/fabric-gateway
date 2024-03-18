# Overview

The _fabric-gateway_ package enables Node.js developers to build client applications using the Hyperledger Fabric programming model as described in the [Running a Fabric Application](https://hyperledger-fabric.readthedocs.io/en/latest/write_first_app.html) tutorial of the Fabric documentation.

[TypeScript](http://www.typescriptlang.org/) definitions are included in this package.

## Getting started

Client applications interact with the blockchain network using a Fabric Gateway. A session for a given client identity is established by calling `connect()` with a gRPC connection to a Fabric Gateway endpoint, client identity, and client signing implementation. The returned `Gateway` enables interaction with any of the blockchain `Networks` (channels) accessible through the Fabric Gateway. This in turn provides access to Smart `Contracts` within chaincode deployed to that blockchain network, and to which transactions can be submitted or queries can be evaluated.

gRPC connections to a Fabric Gateway may be shared by all `Gateway` instances interacting with that Fabric Gateway.

## Example

The following complete example shows how to connect to a Fabric network, submit a transaction and query the ledger state using an instantiated smart contract.

```TypeScript
import * as grpc from '@grpc/grpc-js';
import * as crypto from 'node:crypto';
import { connect, Identity, signers } from '@hyperledger/fabric-gateway';
import { promises as fs } from 'node:fs';
import { TextDecoder } from 'node:util';

const utf8Decoder = new TextDecoder();

async function main(): Promise<void> {
    const credentials = await fs.readFile('path/to/certificate.pem');
    const identity: Identity = { mspId: 'myorg', credentials };

    const privateKeyPem = await fs.readFile('path/to/privateKey.pem');
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    const signer = signers.newPrivateKeySigner(privateKey);

    const tlsRootCert = await fs.readFile('path/to/tlsRootCertificate.pem');
    const client = new grpc.Client('gateway.example.org:1337', grpc.credentials.createSsl(tlsRootCert));

    const gateway = connect({ identity, signer, client });
    try {
        const network = gateway.getNetwork('channelName');
        const contract = network.getContract('chaincodeName');

        const putResult = await contract.submitTransaction('put', 'time', new Date().toISOString());
        console.log('Put result:', utf8Decoder.decode(putResult));

        const getResult = await contract.evaluateTransaction('get', 'time');
        console.log('Get result:', utf8Decoder.decode(getResult));
    } finally {
        gateway.close();
        client.close();
    }
}

main().catch(console.error);
```
