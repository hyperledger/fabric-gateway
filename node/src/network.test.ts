/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Network } from './network';
import { connect, ConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import * as grpc from "@grpc/grpc-js";

describe ('Network', () => {
    let network: Network;
    let client: grpc.Client;

    beforeEach(async () => {
        const identity: Identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        }
        const Client = grpc.makeGenericClientConstructor({}, '');
        client = new Client('example.org:1337', grpc.credentials.createInsecure());
        const options: ConnectOptions = {
            identity,
            client,
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
