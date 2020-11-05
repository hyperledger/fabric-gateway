/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Gateway } from './gateway';
import { Contract } from './contract';
import { Transaction } from './transaction';
import { Signer } from './signer';

jest.mock('./signer');
jest.mock('./transaction');

let contract: Contract;

beforeEach(async () => {
    const signer = new Signer('org1', Buffer.from(''), Buffer.from(''));
    const gw = await Gateway.connect({url: 'test:2001', signer: signer});
    const nw = gw.getNetwork('mychannel');
    contract = nw.getContract('mycontract');

})

test('evaluateTransaction', async () => {
    const result = await contract.evaluateTransaction('txn1', 'arg1', 'arg2');
})

test('submitTransaction', async () => {
    const result = await contract.submitTransaction('txn1', 'arg1', 'arg2');
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