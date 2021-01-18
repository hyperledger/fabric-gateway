/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from './client';
import { Contract } from './contract';
import { connect, Gateway, InternalConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';
import { protos } from './protos/protos';

interface MockGatewayClient extends GatewayClient {
    endorse: jest.Mock<Promise<protos.IPreparedTransaction>, protos.IProposedTransaction[]>,
    evaluate: jest.Mock<Promise<protos.IResult>, protos.IProposedTransaction[]>,
    submit: jest.Mock<Promise<protos.IEvent>, protos.IPreparedTransaction[]>,
}

function newMockGatewayClient(): MockGatewayClient {
    return {
        endorse: jest.fn(),
        evaluate: jest.fn(),
        submit: jest.fn(),
    };
}

describe('Transaction', () => {
    const expectedResult = 'TX_RESULT';

    let client: MockGatewayClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(async () => {
        client = newMockGatewayClient();
        client.endorse.mockResolvedValue({
            envelope: {
                payload: Buffer.from('PAYLOAD'),
            },
            response: {
                value: Buffer.from(expectedResult),
            },
        });

        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        }
        signer = jest.fn().mockResolvedValue('SIGNATURE');

        const options: InternalConnectOptions = {
            identity,
            signer,
            gatewayClient: client,
        };
        gateway = await connect(options);
        network = gateway.getNetwork('CHANNEL_NAME');
        contract = network.getContract('CHAINCODE_ID');
    });

    it('throws on submit error', async () => {
        client.submit.mockRejectedValue(new Error('ERROR_MESSAGE'));

        await expect(contract.submitTransaction('TRANSACTION_NAME'))
            .rejects.toThrow('ERROR_MESSAGE');
    });

    it('returns result', async () => {
        const result = await contract.submitTransaction('TRANSACTION_NAME');

        const actual = Buffer.from(result).toString();
        expect(actual).toBe(expectedResult);
    });

    it('uses signer', async () => {
        signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

        await contract.submitTransaction('TRANSACTION_NAME');

        const preparedTransaction = client.submit.mock.calls[0][0];
        const signature = Buffer.from(preparedTransaction.envelope?.signature ?? '').toString();
        expect(signature).toBe('MY_SIGNATURE');
    });
});
