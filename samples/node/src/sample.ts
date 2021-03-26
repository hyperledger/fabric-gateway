/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {connect, ConnectOptions, Identity, signers} from 'fabric-gateway';
import path from 'path';
import fs from 'fs';
import crypto from 'crypto';
import * as grpc from '@grpc/grpc-js';

const mspId = 'Org1MSP'
const cryptoPath = path.resolve(__dirname, '..', '..', '..', 'scenario', 'fixtures', 'crypto-material', 'crypto-config', 'peerOrganizations', 'org1.example.com');
const certPath = path.resolve(cryptoPath, 'users', 'User1@org1.example.com', 'msp', 'signcerts', 'User1@org1.example.com-cert.pem');
const keyPath = path.resolve(cryptoPath, 'users', 'User1@org1.example.com', 'msp', 'keystore', 'key.pem');
const tlsCertPath = path.resolve(cryptoPath, 'peers', 'peer0.org1.example.com', 'tls', 'ca.crt');
const peerEndpoint = 'localhost:7051'

async function main() {
    const privateKeyPem = await fs.promises.readFile(keyPath);
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    const signer = signers.newPrivateKeySigner(privateKey);

    const certificate = await fs.promises.readFile(certPath);
    const identity: Identity = {
        mspId: mspId,
        credentials: certificate
    };

    const tlsRootCert = fs.readFileSync(tlsCertPath);
    const grpcOptions: Partial<grpc.ChannelOptions> = {
        'grpc.ssl_target_name_override': 'peer0.org1.example.com'
    };
    const GrpcClient = grpc.makeGenericClientConstructor({}, '');
    const client = new GrpcClient(peerEndpoint, grpc.credentials.createSsl(tlsRootCert), grpcOptions)

    const options: ConnectOptions = {
        client: client,
        signer: signer,
        identity: identity,
    };

    const gateway = await connect(options);
    try {
        const network = gateway.getNetwork('mychannel');
        const contract = network.getContract('basic');
        let result = await contract.submitTransaction('put', 'timestamp', (new Date()).toISOString())
        console.log('result = ', result.toString());
        result = await contract.evaluateTransaction('get', 'timestamp');
        console.log('result = ', result.toString());
    } finally {
        gateway.close();
    }
}

main().catch(console.error);
