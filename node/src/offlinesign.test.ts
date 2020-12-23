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
            value: Buffer.from(expectedResult),
        });
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

            await expect(proposal.evaluate())
                .rejects.toThrow();
        });
    
        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), expected);
            await signedProposal.evaluate();
    
            const proposedTransaction = client.evaluate.mock.calls[0][0];
            const actual = Buffer.from(proposedTransaction.proposal?.signature ?? '').toString();
            expect(actual).toBe(expected.toString());
        });
    });

    describe('endorse', () => {
        it('throws with no signer and no explicit signing', async () => {
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await expect(proposal.endorse())
                .rejects.toThrow();
        });
    
        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), expected);
            await signedProposal.endorse();
    
            const proposedTransaction = client.endorse.mock.calls[0][0];
            const actual = Buffer.from(proposedTransaction.proposal?.signature ?? '').toString();
            expect(actual).toBe(expected.toString());
        });
    });

    describe('submit', () => {
        it('throws with no signer and no explicit signing', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const transaction = await signedProposal.endorse();

            expect(transaction.submit())
                .rejects.toThrow();
        });
    
        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), expected);
            await signedTransaction.submit();
    
            const preparedTransaction = client.submit.mock.calls[0][0];
            const actual = Buffer.from(preparedTransaction.envelope?.signature ?? '').toString();
            expect(actual).toBe(expected.toString());
        });
    });
});
