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

describe('Offline sign', () => {
    const expectedResult = 'TX_RESULT';

    let client: MockGatewayClient;
    let identity: Identity;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(async () => {
        client = newMockGatewayClient();
        client.evaluate.mockResolvedValue({
            result: {
                payload: Buffer.from(expectedResult),
            },
        });
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

        const options: InternalConnectOptions = {
            identity,
            gatewayClient: client,
        };
        gateway = await connect(options);
        network = gateway.getNetwork('CHANNEL_NAME');
        contract = network.getContract('CHAINCODE_ID');
    });

    describe('evaluate', () => {
        it('throws with no signer and no explicit signing', async () => {
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await expect(proposal.evaluate()).rejects.toThrow();
        });
    
        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), expected);
            await signedProposal.evaluate();
    
            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const actual = Buffer.from(evaluateRequest.proposed_transaction?.signature ?? '').toString();
            expect(actual).toBe(expected.toString());
        });
    });

    describe('endorse', () => {
        it('throws with no signer and no explicit signing', async () => {
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await expect(proposal.endorse()).rejects.toThrow();
        });
    
        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), expected);
            await signedProposal.endorse();
    
            const endorseRequest = client.endorse.mock.calls[0][0];
            const actual = Buffer.from(endorseRequest.proposed_transaction?.signature ?? '').toString();
            expect(actual).toBe(expected.toString());
        });
    });

    describe('submit', () => {
        it('throws with no signer and no explicit signing', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const transaction = await signedProposal.endorse();

            await expect(transaction.submit()).rejects.toThrow();
        });
    
        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), expected);
            await signedTransaction.submit();
    
            const submitRequest = client.submit.mock.calls[0][0];
            const actual = Buffer.from(submitRequest.prepared_transaction?.signature ?? '').toString();
            expect(actual).toBe(expected.toString());
        });
    });

    describe('serialization', () => {
        it('proposal keeps same transaction ID', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const expected = unsignedProposal.getTransactionId();

            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const actual = signedProposal.getTransactionId();
    
            expect(actual).toBe(expected);
        });

        it('proposal keeps same digest', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const expected = unsignedProposal.getDigest();

            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const actual = signedProposal.getDigest();
    
            expect(actual).toEqual(expected);
        });

        it('transaction keeps same digest', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const expected = unsignedTransaction.getDigest();

            const signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), expected);
            const actual = signedTransaction.getDigest();
    
            expect(actual).toEqual(expected);
        });

        it('transaction keeps same transaction ID', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const digest = unsignedTransaction.getDigest();
            const expected = unsignedTransaction.getTransactionId();

            const signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), digest);
            const actual = signedTransaction.getTransactionId();
    
            expect(actual).toEqual(expected);
        });
    });
});
