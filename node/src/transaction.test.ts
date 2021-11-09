/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions, Metadata, ServiceError } from '@grpc/grpc-js';
import { MockGatewayGrpcClient } from './client.test';
import { CommitError } from './commiterror';
import { CommitStatusError } from './commitstatuserror';
import { Contract } from './contract';
import { Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';
import { Envelope, Status } from './protos/common/common_pb';
import { CommitStatusResponse, EndorseResponse } from './protos/gateway/gateway_pb';
import { Response } from './protos/peer/proposal_response_pb';
import { TxValidationCode } from './protos/peer/transaction_pb';
import { SubmitError } from './submiterror';

describe('Transaction', () => {
    const expectedResult = 'TX_RESULT';
    const serviceError: ServiceError = Object.assign(new Error('ERROR_MESSAGE'), {
        code: Status.SERVICE_UNAVAILABLE,
        details: 'DETAILS',
        metadata: new Metadata(),
    });

    let submitOptions: () => CallOptions;
    let commitStatusOptions: () => CallOptions;
    let client: MockGatewayGrpcClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(() => {
        const now = new Date();
        const submitCallOptions = {
            deadline: now.setHours(now.getHours() + 1),
        }
        submitOptions = () => submitCallOptions; // Return a specific object to test modification
        const commitStatusCallOptions = {
            deadline: now.setHours(now.getHours() + 1),
        }
        commitStatusOptions = () => commitStatusCallOptions; // Return a specific object to test modification

        client = new MockGatewayGrpcClient();

        const txResult = new Response()
        txResult.setPayload(Buffer.from(expectedResult));

        const preparedTx = new Envelope();
        preparedTx.setPayload(Buffer.from('PAYLOAD'));

        const endorseResult = new EndorseResponse();
        endorseResult.setPreparedTransaction(preparedTx);
        endorseResult.setResult(txResult)

        client.mockEndorseResponse(endorseResult);

        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.VALID);

        client.mockCommitStatusResponse(commitResult);

        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        }
        signer = jest.fn(undefined);
        signer.mockResolvedValue(Buffer.from('SIGNATURE'));
        hash = jest.fn(undefined);
        hash.mockReturnValue(Buffer.from('DIGEST'));

        const options: InternalConnectOptions = {
            identity,
            signer,
            hash,
            client,
            submitOptions,
            commitStatusOptions,
        };
        gateway = internalConnect(options);
        network = gateway.getNetwork('CHANNEL_NAME');
        contract = network.getContract('CHAINCODE_NAME');
    });

    it('throws on submit error', async () => {
        client.mockSubmitError(serviceError);
        const transaction = await contract.newProposal('TRANSACTION_NAME').endorse();
        const transactionId = transaction.getTransactionId();

        const t = transaction.submit();

        await expect(t).rejects.toThrow(SubmitError);
        await expect(t).rejects.toThrow(serviceError.message);
        await expect(t).rejects.toMatchObject({
            name: SubmitError.name,
            transactionId,
            cause: serviceError,
        });
    });

    it('throws on commit status error', async () => {
        client.mockCommitStatusError(serviceError);
        const transaction = await contract.newProposal('TRANSACTION_NAME').endorse();
        const commit = await transaction.submit();
        const transactionId = commit.getTransactionId();

        const t = commit.getStatus();

        await expect(t).rejects.toThrow(CommitStatusError);
        await expect(t).rejects.toThrow(serviceError.message);
        await expect(t).rejects.toMatchObject({
            name: CommitStatusError.name,
            transactionId,
            cause: serviceError,
        });
    });

    it('throws CommitError on commit failure', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.MVCC_READ_CONFLICT);
        client.mockCommitStatusResponse(commitResult);

        const t = contract.submitTransaction('TRANSACTION_NAME');

        await expect(t).rejects.toThrow(CommitError);
        await expect(t).rejects.toThrow('MVCC_READ_CONFLICT');
    });

    it('returns result', async () => {
        const result = await contract.submitTransaction('TRANSACTION_NAME');

        const actual = Buffer.from(result).toString();
        expect(actual).toBe(expectedResult);
    });

    it('sets endorsing orgs', async () => {
        await contract.submit('TRANSACTION_NAME', { endorsingOrganizations: ['org1', 'org3'] });
        const actualOrgs = client.getEndorseRequests()[0].getEndorsingOrganizationsList();
        expect(actualOrgs).toStrictEqual(['org1', 'org3']);
    });

    it('includes transaction ID in submit request', async () => {
        await contract.submitTransaction('TRANSACTION_NAME');

        const endorseRequest = client.getEndorseRequests()[0];
        const expected = endorseRequest.getTransactionId();

        const submitRequest = client.getSubmitRequests()[0];
        const actual = submitRequest.getTransactionId();

        expect(actual).toBe(expected);
    });

    it('uses signer for submit', async () => {
        signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

        await contract.submitTransaction('TRANSACTION_NAME');

        const submitRequest = client.getSubmitRequests()[0];
        const signature = Buffer.from(submitRequest.getPreparedTransaction()?.getSignature_asU8() || '').toString();
        expect(signature).toBe('MY_SIGNATURE');
    });

    it('uses signer for commit', async () => {
        signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

        await contract.submitTransaction('TRANSACTION_NAME');

        const statusRequest = client.getCommitStatusRequests()[0];
        const signature = Buffer.from(statusRequest.getSignature() || '').toString();
        expect(signature).toBe('MY_SIGNATURE');
    });

    it('uses hash', async () => {
        hash.mockReturnValue(Buffer.from('MY_DIGEST'));

        await contract.submitTransaction('TRANSACTION_NAME');

        expect(signer).toHaveBeenCalledTimes(3); // endorse, submit and commit
        signer.mock.calls.forEach(call => {
            const digest = call[0].toString();
            expect(digest).toBe('MY_DIGEST');
        });
    });

    it('commit returns transaction validation code', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.MVCC_READ_CONFLICT);
        client.mockCommitStatusResponse(commitResult);

        const commit = await contract.submitAsync('TRANSACTION_NAME');
        const status = await commit.getStatus();

        expect(status.code).toBe(TxValidationCode.MVCC_READ_CONFLICT);
    });

    it('commit returns successful for successful transaction', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.VALID);
        client.mockCommitStatusResponse(commitResult);

        const commit = await contract.submitAsync('TRANSACTION_NAME');
        const status = await commit.getStatus();

        expect(status.successful).toBe(true);
    });

    it('commit returns unsuccessful for failed transaction', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.MVCC_READ_CONFLICT);
        client.mockCommitStatusResponse(commitResult);

        const commit = await contract.submitAsync('TRANSACTION_NAME');
        const status = await commit.getStatus();

        expect(status.successful).toBe(false);
    });

    it('commit returns block number', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.MVCC_READ_CONFLICT);
        commitResult.setBlockNumber(101);
        client.mockCommitStatusResponse(commitResult);

        const commit = await contract.submitAsync('TRANSACTION_NAME');
        const status = await commit.getStatus();

        expect(status.blockNumber).toBe(BigInt(101));
    });

    it('commit returns zero for missing block number', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.MVCC_READ_CONFLICT);
        client.mockCommitStatusResponse(commitResult);

        const commit = await contract.submitAsync('TRANSACTION_NAME');
        const status = await commit.getStatus();

        expect(status.blockNumber).toBe(BigInt(0));
    });

    it('submit uses specified call options', async () => {
        const deadline = Date.now() + 1000;
        const transaction = await contract.newProposal('TRANSACTION_NAME').endorse();

        await transaction.submit({ deadline });

        const actual = client.getSubmitOptions()[0];
        expect(actual.deadline).toBe(deadline);
    });

    it('submit uses default call options', async () => {
        const transaction = await contract.newProposal('TRANSACTION_NAME').endorse();

        await transaction.submit();

        const actual = client.getSubmitOptions()[0];
        expect(actual.deadline).toBe(submitOptions().deadline);
    });

    it('submit default call options are not modified', async () => {
        const expected = submitOptions().deadline;
        const deadline = Date.now() + 1000;
        const transaction = await contract.newProposal('TRANSACTION_NAME').endorse();

        await transaction.submit({ deadline });

        expect(submitOptions().deadline).toBe(expected);
    });

    it('commit uses specified call options', async () => {
        const deadline = Date.now() + 1000;
        const transaction = await contract.newProposal('TRANSACTION_NAME').endorse();
        const commit = await transaction.submit();

        await commit.getStatus({ deadline });

        const actual = client.getCommitStatusOptions()[0];
        expect(actual.deadline).toBe(deadline);
    });

    it('commit uses default call options', async () => {
        const transaction = await contract.newProposal('TRANSACTION_NAME').endorse();
        const commit = await transaction.submit();

        await commit.getStatus();

        const actual = client.getCommitStatusOptions()[0];
        expect(actual.deadline).toBe(commitStatusOptions().deadline);
    });

    it('commit default call options are not modified', async () => {
        const expected = commitStatusOptions().deadline;
        const deadline = Date.now() + 1000;
        const transaction = await contract.newProposal('TRANSACTION_NAME').endorse();
        const commit = await transaction.submit();

        await commit.getStatus({ deadline });

        expect(commitStatusOptions().deadline).toBe(expected);
    });
});
