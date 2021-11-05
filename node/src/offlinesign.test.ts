/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { MockGatewayGrpcClient } from './client.test';
import { Contract } from './contract';
import { Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';
import { ChannelHeader, Envelope, Header, Payload } from './protos/common/common_pb';
import { CommitStatusResponse, EndorseResponse, EvaluateResponse } from './protos/gateway/gateway_pb';
import { Response } from './protos/peer/proposal_response_pb';
import { TxValidationCode } from './protos/peer/transaction_pb';
import { undefinedSignerMessage } from './signingidentity';

describe('Offline sign', () => {
    const expectedResult = 'TX_RESULT';

    let client: MockGatewayGrpcClient;
    let identity: Identity;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(() => {
        client = new MockGatewayGrpcClient();

        const txResult = new Response()
        txResult.setPayload(Buffer.from(expectedResult));

        const evaluateResult = new EvaluateResponse();
        evaluateResult.setResult(txResult)

        client.mockEvaluateResponse(evaluateResult);

        const channelHeader = new ChannelHeader();
        channelHeader.setChannelId('CHANNEL');

        const header = new Header();
        header.setChannelHeader(channelHeader.serializeBinary());

        const payload = new Payload();
        payload.setHeader(header);

        const txEnvelope = new Envelope();
        txEnvelope.setPayload(payload.serializeBinary());

        const endorseResult = new EndorseResponse();
        endorseResult.setPreparedTransaction(txEnvelope);
        endorseResult.setResult(txResult)

        client.mockEndorseResponse(endorseResult);

        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.VALID);

        client.mockCommitStatusResponse(commitResult);

        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        }

        const options: InternalConnectOptions = {
            identity,
            client,
        };
        gateway = internalConnect(options);
        network = gateway.getNetwork('CHANNEL_NAME');
        contract = network.getContract('CHAINCODE_NAME');
    });

    describe('evaluate', () => {
        it('throws with no signer and no explicit signing', async () => {
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await expect(proposal.evaluate()).rejects.toThrow(undefinedSignerMessage);
        });
    
        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), expected);
            await signedProposal.evaluate();
    
            const evaluateRequest = client.getEvaluateRequests()[0];
            const actual = Buffer.from(evaluateRequest.getProposedTransaction()?.getSignature_asU8() || '').toString();
            expect(actual).toBe(expected.toString());
        });

        it('uses offline signature and selected orgs', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME', {endorsingOrganizations: ['org3', 'org5']});
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), expected);
            await signedProposal.evaluate();

            const evaluateRequest = client.getEvaluateRequests()[0];
            const actualOrgs = evaluateRequest.getTargetOrganizationsList();
            expect(actualOrgs).toStrictEqual(['org3', 'org5']);
        });
    });

    describe('endorse', () => {
        it('throws with no signer and no explicit signing', async () => {
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await expect(proposal.endorse()).rejects.toThrow(undefinedSignerMessage);
        });
    
        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), expected);
            await signedProposal.endorse();
    
            const endorseRequest = client.getEndorseRequests()[0];
            const actual = Buffer.from(endorseRequest.getProposedTransaction()?.getSignature_asU8() || '').toString();
            expect(actual).toBe(expected.toString());
        });

        it('uses offline signature and selected orgs', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME', {endorsingOrganizations: ['org3', 'org5']});
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), expected);
            await signedProposal.endorse();

            const endorseRequest = client.getEndorseRequests()[0];
            const actualOrgs = endorseRequest.getEndorsingOrganizationsList();
            expect(actualOrgs).toStrictEqual(['org3', 'org5']);
        });
    });

    describe('submit', () => {
        it('throws with no signer and no explicit signing', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const transaction = await signedProposal.endorse();

            await expect(transaction.submit()).rejects.toThrow(undefinedSignerMessage);
        });
    
        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), expected);
            await signedTransaction.submit();
    
            const submitRequest = client.getSubmitRequests()[0];
            const actual = Buffer.from(submitRequest.getPreparedTransaction()?.getSignature_asU8() || '').toString();
            expect(actual).toBe(expected.toString());
        });
    });

    describe('commit', () => {
        it('throws with no signer and no explicit signing', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), Buffer.from('SIGNATURE'));
            const commit = await signedTransaction.submit();

            await expect(commit.getStatus()).rejects.toThrow(undefinedSignerMessage);
        });

        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedCommit = await signedTransaction.submit();
            const signedCommit = gateway.newSignedCommit(unsignedCommit.getBytes(), expected);
            await signedCommit.getStatus();
    
            const commitRequest = client.getCommitStatusRequests()[0];
            const actual = Buffer.from(commitRequest.getSignature_asU8() ?? '').toString();
            expect(actual).toBe(expected.toString());
        });
    });

    describe('serialization', () => {
        it('proposal keeps same transaction ID', () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const expected = unsignedProposal.getTransactionId();

            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const actual = signedProposal.getTransactionId();
    
            expect(actual).toBe(expected);
        });

        it('proposal keeps same digest', () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const expected = unsignedProposal.getDigest();

            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const actual = signedProposal.getDigest();
    
            expect(actual).toEqual(expected);
        });

        it('transaction keeps same digest', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const expected = unsignedTransaction.getDigest();

            const signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), expected);
            const actual = signedTransaction.getDigest();
    
            expect(actual).toEqual(expected);
        });

        it('transaction keeps same transaction ID', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const expected = unsignedTransaction.getTransactionId();

            const signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), Buffer.from('SIGNATURE'));
            const actual = signedTransaction.getTransactionId();
    
            expect(actual).toEqual(expected);
        });

        it('commit keeps same transaction ID', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedCommit = await signedTransaction.submit();
            const expected = unsignedCommit.getTransactionId();

            const signedCommit = gateway.newSignedCommit(unsignedCommit.getBytes(), Buffer.from('SIGNATURE'))
            const actual = signedCommit.getTransactionId();
    
            expect(actual).toEqual(expected);
        });

        it('commit keeps same digest', async () => {
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedTransaction = await signedProposal.endorse();
            const signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), Buffer.from('SIGNATURE'));
            const unsignedCommit = await signedTransaction.submit();
            const expected = unsignedCommit.getDigest();

            const signedCommit = gateway.newSignedCommit(unsignedCommit.getBytes(), expected)
            const actual = signedCommit.getDigest();
    
            expect(actual).toEqual(expected);
        });
    });
});
