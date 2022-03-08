/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { connect, ConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';

describe('Network', () => {
    let network: Network;
    let client: grpc.Client;

    beforeEach(() => {
        const identity: Identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        };
        client = new grpc.Client('example.org:1337', grpc.credentials.createInsecure());
        const options: ConnectOptions = {
            identity,
            client,
        };

        const gateway = connect(options);
        network = gateway.getNetwork('CHANNEL_NAME');
    });

    describe('getContract', () => {
        it('returns correctly named default contract', () => {
            const contract = network.getContract('CHAINCODE_NAME');

            expect(contract.getChaincodeName()).toBe('CHAINCODE_NAME');
            expect(contract.getContractName()).toBeUndefined();
        });

        it('returns correctly named non-default contract', () => {
            const contract = network.getContract('CHAINCODE_NAME', 'CONTRACT_NAME');

            expect(contract.getChaincodeName()).toBe('CHAINCODE_NAME');
            expect(contract.getContractName()).toBe('CONTRACT_NAME');
        });
    });
});
