# Overview

The *fabric-gateway* package enables Node.js developers to build client applications using the Hyperledger Fabric programming model as described in the [Developing Applications](https://hyperledger-fabric.readthedocs.io/en/latest/developapps/developing_applications.html) chapter of the Fabric documentation.

[TypeScript](http://www.typescriptlang.org/) definitions are included in this package.

## Getting started

Client applications interact with the blockchain network using a Fabric Gateway. A session for a given client identity is established by calling `connect()` with a gRPC connection to a Fabric Gateway endpoint, client identity, and client signing implementation. The returned `Gateway` enables interaction with any of the blockchain `Networks` (channels) accessible through the Fabric Gateway. This in turn provides access to Smart `Contracts` within chaincode deployed to that blockchain network, and to which transactions can be submitted or queries can be evaluated.

gRPC connections to a Fabric Gateway may be shared by all `Gateway` instances interacting with that Fabric Gateway.

## Example

The following complete example shows how to connect to a Fabric network, submit a transaction and query the ledger state using an instantiated smart contract.

    import * as grpc from '@grpc/grpc-js';
    import { connect, Identity, signers } from 'fabric-gateway';

    async main(): void {
        const GrpcClient = grpc.makeGenericClientConstructor({}, '');
        const client = new GrpcClient('gateway.example.org:1337', grpc.credentials.createInsecure());

        const credentials = await fs.promises.readFile('path/to/certificate.pem');
        const identity: Identity = { mspId, credentials };

        const privateKeyPem = await fs.promises.readFile('path/to/privateKey.pem');
        const privateKey = crypto.createPrivateKey(privateKeyPem);
        const signer = signers.newPrivateKeySigner(privateKey);

        const gateway = connect({ client, identity, signer });
        try {
            const network = gateway.getNetwork('channelName');
            const contract = network.getContract('chaincodeName');

            const putResult = await contract.submitTransaction('put', 'time', new Date().toISOString());
            const getResult = await contract.evaluateTransaction('get', 'time');
        } finally {
            gateway.close();
            client.close()
        }
    }

    main().catch(console.log);
