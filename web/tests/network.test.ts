/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { connect, ConnectOptions, Identity, Network } from '../src';

const utf8Encoder = new TextEncoder();

describe('Network', () => {
    let network: Network;

    beforeEach(() => {
        const identity: Identity = {
            mspId: 'MSP_ID',
            credentials: utf8Encoder.encode('CERTIFICATE'),
        };
        const options: ConnectOptions = {
            identity,
            signer: (message) => Promise.resolve(message),
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
