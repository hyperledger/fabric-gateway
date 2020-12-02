/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import * as crypto from 'crypto';
import * as fs from 'fs';
import { connect, InternalConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';
import * as Signers from './identity/signers';
import { Client } from './impl/client';
import { protos } from './protos/protos';
import { Transaction } from './transaction';

let txn: Transaction;
let signer: Signer;
let identity: Identity;

beforeEach(async () => {
    const mspId = 'Org1MSP';
    const certPath = 'test/cert.pem';
    const keyPath = 'test/key.pem';
    const certificate = fs.readFileSync(certPath);
    const keyPem = fs.readFileSync(keyPath);
    const privateKey = crypto.createPrivateKey(keyPem);

    identity = {
        mspId,
        credentials: certificate,
    };
    signer = Signers.newECDSAPrivateKeySigner(privateKey);
    const options = {
        url: 'localhost:7053',
        identity,
        signer,
    };
    const gw = await connect(options);
    const nw = gw.getNetwork('mychannel');
    const contr = nw.getContract('mycontract');
    txn = contr.createTransaction('txn1');
})


test('getName', async () => {
    const name = txn.getName();
    expect(name).toEqual('txn1');
})

test('setTransient', async () => {
    const transient = {
        name1: Buffer.from('value1')
    };
    txn.setTransient(transient);
})

test('create, sign and wrap proposal', async () => {
    const proposal = txn['createProposal'](['arg1', 'arg2']);
    const signed = txn['signProposal'](proposal);
    txn['createProposedWrapper'](signed);
})

const MockClient = class implements Client {
    async _evaluate(signedProposal: protos.IProposedTransaction): Promise<string> { // eslint-disable-line @typescript-eslint/no-unused-vars
        return 'result1';
    }
    async _endorse(signedProposal: protos.IProposedTransaction): Promise<protos.IPreparedTransaction> { // eslint-disable-line @typescript-eslint/no-unused-vars
        return {
            envelope: {
                payload: Buffer.from('payload1')
            },
            response: {
                value: Buffer.from('result2')
            }
        };
    }
    async _submit(preparedTransaction: protos.PreparedTransaction): Promise<protos.IEvent> { // eslint-disable-line @typescript-eslint/no-unused-vars
        return {
        };
    }

}

test('evaluate', async () => {
    const options: InternalConnectOptions = {
        url: 'localhost:7053',
        identity,
        signer,
        client: new MockClient(),
    }
    const gw = await connect(options);
    const nw = gw.getNetwork('mychannel');
    const contr = nw.getContract('mycontract');
    const tx = contr.createTransaction('txn2')
    const result = await tx.evaluate('arg1', 'arg2');
    expect(result).toEqual('result1');
})

test('submit', async () => {
    const options: InternalConnectOptions = {
        url: 'localhost:7053',
        identity,
        signer,
        client: new MockClient(),
    }
    const gw = await connect(options);
    const nw = gw.getNetwork('mychannel');
    const contr = nw.getContract('mycontract');
    const tx = contr.createTransaction('txn2')
    const result = await tx.submit('arg1', 'arg2');
    expect(result).toEqual('result2');
})

