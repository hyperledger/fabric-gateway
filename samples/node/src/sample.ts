/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { ServiceClient } from '@grpc/grpc-js/build/src/make-client';
import * as crypto from 'crypto';
import { connect, Gateway, Identity, Signer, signers } from 'fabric-gateway';
import { promises as fs } from 'fs';
import * as path from 'path';
import { TextDecoder } from 'util';

const mspId = 'Org1MSP'
const cryptoPath = path.resolve(__dirname, '..', '..', '..', 'scenario', 'fixtures', 'crypto-material', 'crypto-config', 'peerOrganizations', 'org1.example.com');
const certPath = path.resolve(cryptoPath, 'users', 'User1@org1.example.com', 'msp', 'signcerts', 'User1@org1.example.com-cert.pem');
const keyPath = path.resolve(cryptoPath, 'users', 'User1@org1.example.com', 'msp', 'keystore', 'key.pem');
const tlsCertPath = path.resolve(cryptoPath, 'peers', 'peer0.org1.example.com', 'tls', 'ca.crt');
const peerEndpoint = 'localhost:7051'

async function main() {
    // The gRPC client connection should be shared by all Gateway connections to this endpoint
    const client = await newGrpcConnection();

    const gateway = connect({
        client,
        identity: await newIdentity(),
        signer: await newSigner(),
    });

    try {
        console.log('exampleSubmit:')
        await exampleSubmit(gateway);
        console.log();

        console.log('exampleSubmitAsync:')
        await exampleSubmitAsync(gateway)
        console.log();

        console.log('exampleSubmitPrivateData:')
        await exampleSubmitPrivateData(gateway)
        console.log();

        console.log('exampleSubmitPrivateData2:')
        await exampleSubmitPrivateData2(gateway)
        console.log();

        console.log('exampleStateBasedEndorsement:')
        await exampleStateBasedEndorsement(gateway)
        console.log();

        console.log('exampleChaincodeEvents:')
        await exampleChaincodeEvents(gateway)
        console.log();
    } finally {
        gateway.close();
        client.close()
        process.exit(0);
    }
}

async function exampleSubmit(gateway: Gateway) {
    const network = gateway.getNetwork('mychannel');
    const contract = network.getContract('basic');

    const timestamp = new Date().toISOString();
    console.log('Submitting "put" transaction with arguments: time,', timestamp);

    // Submit a transaction, blocking until the transaction has been committed on the ledger
    const submitResult = await contract.submitTransaction('put', 'time', timestamp);

    console.log('Submit result:', submitResult.toString());
    console.log('Evaluating "get" query with arguments: time');

    const evaluateResult = await contract.evaluateTransaction('get', 'time');

    console.log('Query result:', evaluateResult.toString());
}

async function exampleSubmitAsync(gateway: Gateway) {
    const network = gateway.getNetwork('mychannel');
    const contract = network.getContract('basic');

    const timestamp = new Date().toISOString();
    console.log('Submitting "put" transaction asynchronously with arguments: async,', timestamp);

	// Submit transaction asynchronously, blocking until the transaction has been sent to the orderer, and allowing
	// this thread to process the chaincode response (e.g. update a UI) without waiting for the commit notification
    const commit = await contract.submitAsync('put', {
        arguments: ['async', timestamp],
    });
    const submitResult = commit.getResult();

    console.log('Submit result:', submitResult.toString());
    console.log('Waiting for transaction commit');

    const successful = await commit.isSuccessful();
    if (!successful) {
        const status = await commit.getStatus();
        throw new Error(`Transaction ${commit.getTransactionId()} failed to commit with status code: ${status}`);
    }

    console.log('Transaction committed successfully');
    console.log('Evaluating "get" query with arguments: async');

    const evaluateResult = await contract.evaluateTransaction('get', 'async');

    console.log('Query result:', evaluateResult.toString());
}

async function exampleSubmitPrivateData(gateway: Gateway) {
    const network = gateway.getNetwork('mychannel');
    const contract = network.getContract('private');

    const timestamp = new Date().toISOString();
    const privateData = {
        'collection': Buffer.from('SharedCollection,Org3Collection'),
        'key': Buffer.from('my-private-key'),
        'value': Buffer.from(timestamp)
    }
    console.log('Submitting "WritePrivateData" transaction with private data: ', privateData.value.toString());

    // Submit transaction, blocking until the transaction has been committed on the ledger.
    // The 'transient' data will not get written to the ledger, and is used to send sensitive data to the trusted endorsing peers.
    // The gateway will only send this to peers that are included in the ownership policy of all collections accessed by the chaincode function.
    // It is assumed that the gateway's organization is trusted and will invoke the chaincode to work out if extra endorsements are required from other orgs.
    // In this example, it will also seek endorsement from Org3, which is included in the ownership policy of both collections.
    await contract.submit('WritePrivateData', { transientData: privateData });

    console.log('Evaluating "ReadPrivateData" query with arguments: "SharedCollection", "my-private-key"');
    const evaluateResult = await contract.evaluateTransaction('ReadPrivateData', 'SharedCollection', 'my-private-key');
    console.log('Query result:', evaluateResult.toString());
}

async function exampleSubmitPrivateData2(gateway: Gateway) {
    const network = gateway.getNetwork('mychannel');
    const contract = network.getContract('private');

    const timestamp = new Date().toISOString();
    const privateData = {
        'collection': Buffer.from('Org1Collection,Org3Collection'),
        'key': Buffer.from('my-private-key2'),
        'value': Buffer.from(timestamp)
    }
    console.log('Submitting "WritePrivateData" transaction with private data: ', privateData.value.toString());

    // This example is similar to the previous private data example.
    // The difference here is that the gateway cannot assume that Org3 is trusted to receive transient data
    // that might be destined for storage in Org1Collection, since Org3 is not in its ownership policy.
    // The client application must explicitly specify which organizations must endorse using the endorsingOrganizations option.
    await contract.submit('WritePrivateData', { transientData: privateData, endorsingOrganizations: ['Org1MSP', 'Org3MSP'] });

    console.log('Evaluating "ReadPrivateData" query with arguments: "Org1Collection", "my-private-key2"');
    const evaluateResult = await contract.evaluateTransaction('ReadPrivateData', 'Org1Collection', 'my-private-key2');
    console.log('Query result:', evaluateResult.toString());
}

async function exampleStateBasedEndorsement(gateway: Gateway) {
    const network = gateway.getNetwork('mychannel');
    const contract = network.getContract('private');

    console.log('Submitting "SetStateWithEndorser" transaction with arguments:  "sbe-key", "value1", "Org1MSP"');
    // Submit a transaction, blocking until the transaction has been committed on the ledger
    await contract.submitTransaction('SetStateWithEndorser', 'sbe-key', 'value1', 'Org1MSP');

    // Query the current state
    console.log('Evaluating "GetState" query with arguments: "sbe-key"');
    let evaluateResult = await contract.evaluateTransaction('GetState', 'sbe-key');
    console.log('Query result:', evaluateResult.toString());

    // Submit transaction to modify the state.
    // The state-based endorsement policy will override the chaincode policy for this state (key).
    console.log('Submitting "ChangeState" transaction with arguments:  "sbe-key", "value2"');
    await contract.submitTransaction('ChangeState', 'sbe-key', 'value2');

    // Verify the current state
    console.log('Evaluating "GetState" query with arguments: "sbe-key"');
    evaluateResult = await contract.evaluateTransaction('GetState', 'sbe-key');
    console.log('Query result:', evaluateResult.toString());

    // Now change the state-based endorsement policy for this state.
    console.log('Submitting "SetStateEndorsers" transaction with arguments:  "sbe-key", "Org2MSP", "Org3MSP"');
    await contract.submitTransaction('SetStateEndorsers', 'sbe-key', 'Org2MSP', 'Org3MSP');

    // Modify the state.  It will now require endorsement from Org2 and Org3 for this transaction to succeed.
    // The gateway will endorse this transaction proposal on one of its organization's peers and will determine if
    // extra endorsements are required to satisfy any state changes.
    // In this example, it will seek endorsements from Org2 and Org3 in order to satisfy the SBE policy.
    console.log('Submitting "ChangeState" transaction with arguments:  "sbe-key", "value3"');
    await contract.submitTransaction('ChangeState', 'sbe-key', 'value3');

    // Verify the new state
    console.log('Evaluating "GetState" query with arguments: "sbe-key"');
    evaluateResult = await contract.evaluateTransaction('GetState', 'sbe-key');
    console.log('Query result:', evaluateResult.toString());
}

async function exampleChaincodeEvents(gateway: Gateway) {
    const network = gateway.getNetwork('mychannel');
    const contract = network.getContract('basic');

    console.log('Listening for chaincode events');
    const events = await network.getChaincodeEvents('basic');

    // Submit a transaction that generates a chaincode event
    console.log('Submitting "event" transaction with arguments:  "my-event-name", "my-event-payload"');
    await contract.submitTransaction('event', 'my-event-name', 'my-event-payload');

    const decoder = new TextDecoder();
    for await (const event of events) {
        const payload = decoder.decode(event.payload);
        console.log(`Received event name: ${event.eventName}, payload: ${payload}, txID: ${event.transactionId}`);
        break;
    }
}

async function newGrpcConnection(): Promise<ServiceClient> {
    const tlsRootCert = await fs.readFile(tlsCertPath);
    const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);

    const GrpcClient = grpc.makeGenericClientConstructor({}, '');
    return new GrpcClient(peerEndpoint, tlsCredentials, {
        'grpc.ssl_target_name_override': 'peer0.org1.example.com'
    });
}

async function newIdentity(): Promise<Identity> {
    const credentials = await fs.readFile(certPath);
    return { mspId, credentials };
}

async function newSigner(): Promise<Signer> {
    const privateKeyPem = await fs.readFile(keyPath);
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    return signers.newPrivateKeySigner(privateKey);
}

main().catch(console.error);
