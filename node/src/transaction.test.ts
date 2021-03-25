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
import { gateway, protos } from './protos/protos';

interface MockGatewayClient extends GatewayClient {
    endorse: jest.Mock<Promise<gateway.IEndorseResponse>, gateway.IEndorseRequest[]>,
    evaluate: jest.Mock<Promise<gateway.IEvaluateResponse>, gateway.IEvaluateRequest[]>,
    submit: jest.Mock<Promise<gateway.ISubmitResponse>, gateway.ISubmitRequest[]>,
    commitStatus: jest.Mock<Promise<gateway.ICommitStatusResponse>, gateway.ICommitStatusRequest[]>,
}

function newMockGatewayClient(): MockGatewayClient {
    return {
        endorse: jest.fn(),
        evaluate: jest.fn(),
        submit: jest.fn(),
        commitStatus: jest.fn(),
    };
}

describe('Transaction', () => {
    const expectedResult = 'TX_RESULT';

    let client: MockGatewayClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(async () => {
        client = newMockGatewayClient();
        client.endorse.mockResolvedValue({
            prepared_transaction: {
                payload: Buffer.from('PAYLOAD'),
            },
            result: {
                payload: Buffer.from(expectedResult),
            },
        });
        client.commitStatus.mockResolvedValue({
            result: protos.TxValidationCode.VALID,
        });

        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        }
        signer = jest.fn().mockResolvedValue(Buffer.from('SIGNATURE'));
        hash = jest.fn().mockReturnValue(Buffer.from('DIGEST'));

        const options: InternalConnectOptions = {
            identity,
            signer,
            hash,
            gatewayClient: client,
        };
        gateway = await connect(options);
        network = gateway.getNetwork('CHANNEL_NAME');
        contract = network.getContract('CHAINCODE_ID');
    });

    it('throws on submit error', async () => {
        client.submit.mockRejectedValue(new Error('ERROR_MESSAGE'));

        await expect(contract.submitTransaction('TRANSACTION_NAME')).rejects.toThrow('ERROR_MESSAGE');
    });

    it('throws on commit failure', async () => {
        client.commitStatus.mockResolvedValue({
            result: protos.TxValidationCode.MVCC_READ_CONFLICT,
        });

        await expect(contract.submitTransaction('TRANSACTION_NAME'))
            .rejects.toThrow(protos.TxValidationCode[protos.TxValidationCode.MVCC_READ_CONFLICT]);
    });

    it('returns result', async () => {
        const result = await contract.submitTransaction('TRANSACTION_NAME');

        const actual = Buffer.from(result).toString();
        expect(actual).toBe(expectedResult);
    });

    it('uses signer', async () => {
        signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

        await contract.submitTransaction('TRANSACTION_NAME');

        const submitRequest = client.submit.mock.calls[0][0];
        const signature = Buffer.from(submitRequest.prepared_transaction?.signature ?? '').toString();
        expect(signature).toBe('MY_SIGNATURE');
    });

    it('uses hash', async () => {
        hash.mockReturnValue(Buffer.from('MY_DIGEST'));

        await contract.submitTransaction('TRANSACTION_NAME');

        expect(signer).toHaveBeenCalledTimes(2); // endorse and submit
        signer.mock.calls.forEach(call => {
            const digest = call[0].toString();
            expect(digest).toBe('MY_DIGEST');
        });
    });
});
