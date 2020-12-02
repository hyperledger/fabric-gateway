/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Contract } from './contract';
import { connect } from './gateway';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';

const identity: Identity = {
    mspId: 'MSP_ID',
    credentials: Buffer.from('CERTIFICATE'),
}
const signer: Signer = () => Uint8Array.of();

test('getContract', async () => {
    const gw = await connect({url: 'test:2001', identity, signer});
    const nw = gw.getNetwork('mychannel');
    const contr: Contract = nw.getContract('mycontract');
    expect(contr.getChaincodeId()).toBe('mycontract');
})
