/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { MockGatewayClient, newMockGatewayClient } from './client.test';
import { CommitError } from './commiterror';
import { Contract } from './contract';
import { Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';
import { Envelope } from './protos/common/common_pb';
import { CommitStatusResponse, EndorseResponse } from './protos/gateway/gateway_pb';
import { Response } from './protos/peer/proposal_response_pb';
import { TxValidationCode } from './protos/peer/transaction_pb';

describe('Transaction', () => {
    const expectedResult = 'TX_RESULT';

    let client: MockGatewayClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(() => {
        client = newMockGatewayClient();

        const txResult = new Response()
        txResult.setPayload(Buffer.from(expectedResult));

        const preparedTx = new Envelope();
        preparedTx.setPayload(Buffer.from('PAYLOAD'));

        const endorseResult = new EndorseResponse();
        endorseResult.setPreparedTransaction(preparedTx);
        endorseResult.setResult(txResult)

        client.endorse.mockResolvedValue(endorseResult);

        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.VALID);

        client.commitStatus.mockResolvedValue(commitResult);

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
            gatewayClient: client,
        };
        gateway = internalConnect(options);
        network = gateway.getNetwork('CHANNEL_NAME');
        contract = network.getContract('CHAINCODE_ID');
    });

    it('throws on submit error', async () => {
        client.submit.mockRejectedValue(new Error('ERROR_MESSAGE'));

        await expect(contract.submitTransaction('TRANSACTION_NAME')).rejects.toThrow('ERROR_MESSAGE');
    });

    it('throws CommitError on commit failure', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.MVCC_READ_CONFLICT);
        client.commitStatus.mockResolvedValue(commitResult);

        const t = contract.submitTransaction('TRANSACTION_NAME');

        await expect(t).rejects.toThrow('MVCC_READ_CONFLICT');
        await expect(t).rejects.toThrow(CommitError);
    });

    it('returns result', async () => {
        const result = await contract.submitTransaction('TRANSACTION_NAME');

        const actual = Buffer.from(result).toString();
        expect(actual).toBe(expectedResult);
    });

    it('sets endorsing orgs', async () => {
        await contract.submit('TRANSACTION_NAME', { endorsingOrganizations: ['org1', 'org3']});
        const actualOrgs = client.endorse.mock.calls[0][0].getEndorsingOrganizationsList();
        expect(actualOrgs).toStrictEqual(['org1', 'org3']);
    });

    it('includes transaction ID in submit request', async () => {
        await contract.submitTransaction('TRANSACTION_NAME');

        const endorseRequest = client.endorse.mock.calls[0][0];
        const expected = endorseRequest.getTransactionId();

        const submitRequest = client.submit.mock.calls[0][0];
        const actual = submitRequest.getTransactionId();

        expect(actual).toBe(expected);
    });

    it('uses signer for submit', async () => {
        signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

        await contract.submitTransaction('TRANSACTION_NAME');

        const submitRequest = client.submit.mock.calls[0][0];
        const signature = Buffer.from(submitRequest.getPreparedTransaction()?.getSignature_asU8() || '').toString();
        expect(signature).toBe('MY_SIGNATURE');
    });

    it('uses signer for commit', async () => {
        signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

        await contract.submitTransaction('TRANSACTION_NAME');

        const statusRequest = client.commitStatus.mock.calls[0][0];
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
        client.commitStatus.mockResolvedValue(commitResult);

        const commit = await contract.submitAsync('TRANSACTION_NAME');
        const status = await commit.getStatus();

        expect(status.code).toBe(TxValidationCode.MVCC_READ_CONFLICT);
    });

    it('commit returns successful for successful transaction', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.VALID);
        client.commitStatus.mockResolvedValue(commitResult);

        const commit = await contract.submitAsync('TRANSACTION_NAME');
        const status = await commit.getStatus();

        expect(status.successful).toBe(true);
    });

    it('commit returns unsuccessful for failed transaction', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.MVCC_READ_CONFLICT);
        client.commitStatus.mockResolvedValue(commitResult);

        const commit = await contract.submitAsync('TRANSACTION_NAME');
        const status = await commit.getStatus();

        expect(status.successful).toBe(false);
    });

    it('commit returns block number', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.MVCC_READ_CONFLICT);
        commitResult.setBlockNumber(101);
        client.commitStatus.mockResolvedValue(commitResult);

        const commit = await contract.submitAsync('TRANSACTION_NAME');
        const status = await commit.getStatus();

        expect(status.blockNumber).toBe(BigInt(101));
    });

    it('commit returns zero for missing block number', async () => {
        const commitResult = new CommitStatusResponse();
        commitResult.setResult(TxValidationCode.MVCC_READ_CONFLICT);
        client.commitStatus.mockResolvedValue(commitResult);

        const commit = await contract.submitAsync('TRANSACTION_NAME');
        const status = await commit.getStatus();

        expect(status.blockNumber).toBe(BigInt(0));
    });
});
