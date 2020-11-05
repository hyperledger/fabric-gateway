/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Gateway } from './gateway';
import { Contract } from './contract';
import { Signer } from './signer';

jest.mock('./signer');

test('getContract', async () => {
    const signer = new Signer('org1', Buffer.from(''), Buffer.from(''));
    const gw = await Gateway.connect({url: 'test:2001', signer: signer});
    const nw = gw.getNetwork('mychannel');
    const contr = nw.getContract('mycontract');
    expect(contr).toBeInstanceOf(Contract);
    expect(contr.getName()).toBe('mycontract');
})