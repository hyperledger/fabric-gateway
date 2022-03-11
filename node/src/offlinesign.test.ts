/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Contract } from './contract';
import { Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';
import { Envelope } from './protos/common/common_pb';
import { CommitStatusResponse, EvaluateResponse } from './protos/gateway/gateway_pb';
import { DeliverResponse } from './protos/peer/events_pb';
import { Response } from './protos/peer/proposal_response_pb';
import { TxValidationCode } from './protos/peer/transaction_pb';
import { undefinedSignerMessage } from './signingidentity';
import { MockGatewayGrpcClient, newDuplexStreamResponse, newEndorseResponse } from './testutils.test';

describe('Offline sign', () => {
    const expectedResult = 'TX_RESULT';

    let client: MockGatewayGrpcClient;
    let identity: Identity;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(() => {
        client = new MockGatewayGrpcClient();

        const txResult = new Response();
        txResult.setPayload(Buffer.from(expectedResult));

        const evaluateResult = new EvaluateResponse();
        evaluateResult.setResult(txResult);

        client.mockEvaluateResponse(evaluateResult);

        const endorseResponse = newEndorseResponse({
            result: Buffer.from(expectedResult),
            channelName: 'CHANNEL',
        });
        client.mockEndorseResponse(endorseResponse);

        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.VALID);

        client.mockCommitStatusResponse(commitResult);

        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        };

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
            const actual = Buffer.from(evaluateRequest.getProposedTransaction()?.getSignature_asU8() ?? '').toString();
            expect(actual).toBe(expected.toString());
        });

        it('retains endorsing orgs', async () => {
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
            const actual = Buffer.from(endorseRequest.getProposedTransaction()?.getSignature_asU8() ?? '').toString();
            expect(actual).toBe(expected.toString());
        });

        it('retains endorsing orgs', async () => {
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
            const actual = Buffer.from(submitRequest.getPreparedTransaction()?.getSignature_asU8() ?? '').toString();
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

    describe('chaincode events', () => {
        it('throws with no signer and no explicit signing', async () => {
            const unsignedRequest = network.newChaincodeEventsRequest('CHAINCODE_NAME');

            await expect(unsignedRequest.getEvents()).rejects.toThrow(undefinedSignerMessage);
        });

        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');

            const unsignedRequest = network.newChaincodeEventsRequest('CHAINCODE_NAME');
            const signedRequest = gateway.newSignedChaincodeEventsRequest(unsignedRequest.getBytes(), expected);
            await signedRequest.getEvents();

            const eventsRequest = client.getChaincodeEventsRequests()[0];
            const actual = Buffer.from(eventsRequest.getSignature_asU8() ?? '').toString();
            expect(actual).toBe(expected.toString());
        });
    });

    describe('block events', () => {
        it('throws with no signer and no explicit signing', async () => {
            const unsignedRequest = network.newBlockEventsRequest();

            await expect(unsignedRequest.getEvents()).rejects.toThrow(undefinedSignerMessage);
        });

        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');
            const stream = newDuplexStreamResponse<Envelope, DeliverResponse>([]);
            client.mockBlockEventsResponse(stream);

            const unsignedRequest = network.newBlockEventsRequest();
            const signedRequest = gateway.newSignedBlockEventsRequest(unsignedRequest.getBytes(), expected);
            await signedRequest.getEvents();

            expect(stream.write.mock.calls.length).toBe(1);
            const eventsRequest = stream.write.mock.calls[0][0];
            const actual = Buffer.from(eventsRequest.getSignature_asU8() ?? '').toString();
            expect(actual).toBe(expected.toString());
        });
    });

    describe('filtered block events', () => {
        it('throws with no signer and no explicit signing', async () => {
            const unsignedRequest = network.newFilteredBlockEventsRequest();

            await expect(unsignedRequest.getEvents()).rejects.toThrow(undefinedSignerMessage);
        });

        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');
            const stream = newDuplexStreamResponse<Envelope, DeliverResponse>([]);
            client.mockFilteredBlockEventsResponse(stream);

            const unsignedRequest = network.newFilteredBlockEventsRequest();
            const signedRequest = gateway.newSignedFilteredBlockEventsRequest(unsignedRequest.getBytes(), expected);
            await signedRequest.getEvents();

            expect(stream.write.mock.calls.length).toBe(1);
            const eventsRequest = stream.write.mock.calls[0][0];
            const actual = Buffer.from(eventsRequest.getSignature_asU8() ?? '').toString();
            expect(actual).toBe(expected.toString());
        });
    });

    describe('block events with private data', () => {
        it('throws with no signer and no explicit signing', async () => {
            const unsignedRequest = network.newBlockEventsWithPrivateDataRequest();

            await expect(unsignedRequest.getEvents()).rejects.toThrow(undefinedSignerMessage);
        });

        it('uses offline signature', async () => {
            const expected = Buffer.from('MY_SIGNATURE');
            const stream = newDuplexStreamResponse<Envelope, DeliverResponse>([]);
            client.mockBlockEventsWithPrivateDataResponse(stream);

            const unsignedRequest = network.newBlockEventsWithPrivateDataRequest();
            const signedRequest = gateway.newSignedBlockEventsWithPrivateDataRequest(unsignedRequest.getBytes(), expected);
            await signedRequest.getEvents();

            expect(stream.write.mock.calls.length).toBe(1);
            const eventsRequest = stream.write.mock.calls[0][0];
            const actual = Buffer.from(eventsRequest.getSignature_asU8() ?? '').toString();
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

            const signedCommit = gateway.newSignedCommit(unsignedCommit.getBytes(), Buffer.from('SIGNATURE'));
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

            const signedCommit = gateway.newSignedCommit(unsignedCommit.getBytes(), expected);
            const actual = signedCommit.getDigest();

            expect(actual).toEqual(expected);
        });

        it('chaincode events request keeps same digest', () => {
            const unsignedRequest = network.newChaincodeEventsRequest('CHAINCODE_NAME');
            const expected = unsignedRequest.getDigest();

            const signedRequest = gateway.newSignedChaincodeEventsRequest(unsignedRequest.getBytes(), expected);
            const actual = signedRequest.getDigest();

            expect(actual).toEqual(expected);
        });

        it('block events request keeps same digest', () => {
            const unsignedRequest = network.newBlockEventsRequest();
            const expected = unsignedRequest.getDigest();

            const signedRequest = gateway.newSignedBlockEventsRequest(unsignedRequest.getBytes(), expected);
            const actual = signedRequest.getDigest();

            expect(actual).toEqual(expected);
        });

        it('filtered block events request keeps same digest', () => {
            const unsignedRequest = network.newFilteredBlockEventsRequest();
            const expected = unsignedRequest.getDigest();

            const signedRequest = gateway.newSignedFilteredBlockEventsRequest(unsignedRequest.getBytes(), expected);
            const actual = signedRequest.getDigest();

            expect(actual).toEqual(expected);
        });

        it('block events with private data request keeps same digest', () => {
            const unsignedRequest = network.newBlockEventsWithPrivateDataRequest();
            const expected = unsignedRequest.getDigest();

            const signedRequest = gateway.newSignedBlockEventsWithPrivateDataRequest(unsignedRequest.getBytes(), expected);
            const actual = signedRequest.getDigest();

            expect(actual).toEqual(expected);
        });
    });
});
