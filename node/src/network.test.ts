/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Network } from 'network';
import { connect, ConnectOptions } from './gateway';
import { Identity } from './identity/identity';

describe ('Network', () => {
    let network: Network;

    beforeEach(async () => {
        const identity: Identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        }
        const options: ConnectOptions = {
            identity,
            url: 'example.org:1337',
        };

        const gateway = await connect(options);
        network = gateway.getNetwork('CHANNEL_NAME');
    });

    describe('getContract', () => {
        it('returns correctly named default contract', async () => {
            const contract = network.getContract('CHAINCODE_ID');

            expect(contract.getChaincodeId()).toBe('CHAINCODE_ID');
            expect(contract.getContractName()).toBeUndefined();
        });

        it('returns correctly named non-default contract', async () => {
            const contract = network.getContract('CHAINCODE_ID', 'CONTRACT_NAME');

            expect(contract.getChaincodeId()).toBe('CHAINCODE_ID');
            expect(contract.getContractName()).toBe('CONTRACT_NAME');
        });
    });
});
