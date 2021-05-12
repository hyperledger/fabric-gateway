/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { ServiceClient } from '@grpc/grpc-js/build/src/make-client';
import crypto from 'crypto';
import { connect, Gateway, Identity, Signer, signers } from 'fabric-gateway';
import fs from 'fs';
import path from 'path';

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
        await exampleSubmit(gateway);
        console.log();

        await exampleSubmitAsync(gateway)
        console.log();
    } finally {
        gateway.close();
        client.close()
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

async function newGrpcConnection(): Promise<ServiceClient> {
    const tlsRootCert = await fs.promises.readFile(tlsCertPath);
    const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);

    const GrpcClient = grpc.makeGenericClientConstructor({}, '');
    return new GrpcClient(peerEndpoint, tlsCredentials, {
        'grpc.ssl_target_name_override': 'peer0.org1.example.com'
    });
}

async function newIdentity(): Promise<Identity> {
    const certificate = await fs.promises.readFile(certPath);
    return {
        mspId: mspId,
        credentials: certificate
    };
}

async function newSigner(): Promise<Signer> {
    const privateKeyPem = await fs.promises.readFile(keyPath);
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    return signers.newPrivateKeySigner(privateKey);
}

main().catch(console.error);
