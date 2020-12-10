/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Network } from 'network';
import { connect, Gateway } from './gateway';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';

const identity: Identity = {
    mspId: 'MSP_ID',
    credentials: Buffer.from('CERTIFICATE'),
}
const signer: Signer = () => Uint8Array.of();

test('connect to gateway', async () => {
    const gw: Gateway = await connect({url: 'test:2001', identity, signer});
    expect(gw).toBeDefined();
})

test('getNetwork', async () => {
    const gw = await connect({url: 'test:2001', identity, signer});
    const nw: Network = gw.getNetwork('mychannel');
    expect(nw).toBeDefined();
    expect(nw.getName()).toBe('mychannel')
})