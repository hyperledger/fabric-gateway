/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Contract } from './contract';
import { connect } from './gateway';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';
import { Transaction } from './transaction';

jest.mock('./transaction');

let contract: Contract;

beforeEach(async () => {
    const identity: Identity = {
        mspId: 'MSP_ID',
        credentials: Buffer.from('CERTIFICATE'),
    }
    const signer: Signer = () => Uint8Array.of();
    const gw = await connect({ url: 'test:2001', identity, signer });
    const nw = gw.getNetwork('mychannel');
    contract = nw.getContract('mycontract');

})

test('evaluateTransaction', async () => {
    await contract.evaluateTransaction('txn1', 'arg1', 'arg2');
})

test('submitTransaction', async () => {
    await contract.submitTransaction('txn1', 'arg1', 'arg2');
})

test('createTransaction', () => {
    const txn = contract.createTransaction('txn1');
    expect(txn).toBeInstanceOf(Transaction);
})

test('prepareToEvaluate', async () => {
    const txn = contract.prepareToEvaluate('txn1');
    txn.setArgs('arg1', 'arg2');
    await txn.invoke();
})

test('prepareToSubmit', async () => {
    const txn = contract.prepareToSubmit('txn1');
    txn.setArgs('arg1', 'arg2');
    await txn.invoke();
})