/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Gateway } from './gateway';
import { Signer } from './signer';
import { Transaction } from './transaction';
import { Client } from './impl/client';
import { protos } from './protos/protos'
import * as fs from 'fs';

let txn: Transaction;
let signer: Signer;

beforeEach(async () => {
    const mspid = 'Org1MSP';
    const certPath = 'test/cert.pem';
    const keyPath = 'test/key.pem';
    const cert = fs.readFileSync(certPath);
    const key = fs.readFileSync(keyPath);

    signer = new Signer(mspid, cert, key);
    const options = {
        url: 'localhost:7053',
        signer: signer
    }
    const gw = await Gateway.connect(options);
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
    const proposal = txn['createProposal'](['arg1', 'arg2'], signer);
    const signed = txn['signProposal'](proposal, signer);
    const wrapped = txn['createProposedWrapper'](signed);
})

const MockClient = class implements Client {
    async _evaluate(signedProposal: protos.IProposedTransaction): Promise<string> {
        return 'result1';
    }
    async _endorse(signedProposal: protos.IProposedTransaction): Promise<protos.IPreparedTransaction> {
        return {
            envelope: {
                payload: Buffer.from('payload1')
            },
            response: {
                value: Buffer.from('result2')
            }
        };
    }
    async _submit(preparedTransaction: protos.PreparedTransaction): Promise<protos.IEvent> {
        return {
        };
    }

}

test('evaluate', async () => {
    const options = {
        url: 'localhost:7053',
        signer: signer
    }
    const gw = await Gateway.connect(options);
    gw._client = new MockClient();
    const nw = gw.getNetwork('mychannel');
    const contr = nw.getContract('mycontract');
    const tx = contr.createTransaction('txn2')
    const result = await tx.evaluate('arg1', 'arg2');
    expect(result).toEqual('result1');
})

test('submit', async () => {
    const options = {
        url: 'localhost:7053',
        signer: signer
    }
    const gw = await Gateway.connect(options);
    gw._client = new MockClient();
    const nw = gw.getNetwork('mychannel');
    const contr = nw.getContract('mycontract');
    const tx = contr.createTransaction('txn2')
    const result = await tx.submit('arg1', 'arg2');
    expect(result).toEqual('result2');
})

