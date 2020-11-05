/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Gateway } from './gateway';
import { Network } from './network';
import { Signer } from './signer';

jest.mock('./signer');
  
test('connect to gateway', async () => {
    const signer = new Signer("org1", Buffer.from(''), Buffer.from(''));
    const gw = await Gateway.connect({url: 'test:2001', signer: signer});
    expect(gw).toBeInstanceOf(Gateway);
})

test('getNetwork', async () => {
    const signer = new Signer('org1', Buffer.from(''), Buffer.from(''));
    const gw = await Gateway.connect({url: 'test:2001', signer: signer});
    const nw = gw.getNetwork('mychannel');
    expect(nw).toBeInstanceOf(Network);
    expect(nw.getName()).toBe('mychannel')
})