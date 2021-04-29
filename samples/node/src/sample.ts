/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import * as grpc from '@grpc/grpc-js';
import crypto from 'crypto';
import { connect, ConnectOptions, Contract, Identity, signers } from 'fabric-gateway';
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
    const client = newGrpcConnection();

    const options: ConnectOptions = {
        client,
        identity: await newIdentity(),
        signer: await newSigner(),
    };
    const gateway = await connect(options);

    try {
        const network = gateway.getNetwork('mychannel');
        const contract = network.getContract('basic');

        await exampleSubmit(contract, 'put', 'timestamp', new Date().toISOString());
        await exampleEvaluate(contract, 'get', 'timestamp');

        console.log();

        await exampleSubmitAsync(contract, 'put', 'async', new Date().toISOString())
        await exampleEvaluate(contract, 'get', 'async');

        console.log();
    } finally {
        gateway.close();
        client.close()
    }
}

async function exampleSubmit(contract: Contract, name: string, ...args: string[]) {
    console.log(`Submitting "${name}" transaction with arguments:`, args);

    // Submit a transaction, blocking until the transaction has been committed on the ledger.
    const submitResult = await contract.submitTransaction(name, ...args);
    console.log('Submit result:', submitResult.toString());
}

async function exampleSubmitAsync(contract: Contract, name: string, ...args: string[]) {
    console.log(`Submitting "${name}" transaction asynchronously with arguments:`, args);

	// Submit transaction asynchronously, blocking until the transaction has been sent to the orderer, and allowing
	// this thread to process the chaincode response (e.g. update a UI) without waiting for the commit notification
    const commit = await contract.submitAsync(name, { arguments: args });
    const result = commit.getResult();
    console.log('Proposal result:', result.toString());

    console.log('Waiting for transaction commit');

    const successful = await commit.isSuccessful();
    if (!successful) {
        const status = await commit.getStatus();
        throw new Error(`Transaction ${commit.getTransactionId()} failed to commit with status code ${status}`)
    }
    console.log('Transaction committed successfully')
}

async function exampleEvaluate(contract: Contract, name: string, ...args: string[]) {
    console.log(`Evaluating "${name}" query with arguments: ${args}`);

    const result = await contract.evaluateTransaction('get', 'timestamp');
    console.log('Query result:', result.toString());
}

function newGrpcConnection() {
    const tlsRootCert = fs.readFileSync(tlsCertPath);
    const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);

    const GrpcClient = grpc.makeGenericClientConstructor({}, '');
    return new GrpcClient(peerEndpoint, tlsCredentials, {
        'grpc.ssl_target_name_override': 'peer0.org1.example.com'
    });
}

async function newIdentity() {
    const certificate = await fs.promises.readFile(certPath);
    const identity: Identity = {
        mspId: mspId,
        credentials: certificate
    };
    return identity;
}

async function newSigner() {
    const privateKeyPem = await fs.promises.readFile(keyPath);
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    return signers.newPrivateKeySigner(privateKey);
}

main().catch(console.error);
