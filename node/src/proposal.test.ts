/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions, Metadata, ServiceError, status } from '@grpc/grpc-js';
import { common, gateway as gatewayproto, msp, peer } from '@hyperledger/fabric-protos';
import { Contract } from './contract';
import { EndorseError } from './endorseerror';
import { Gateway, internalConnect } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';
import { MockGatewayGrpcClient, newEndorseResponse } from './testutils.test';

function assertDecodeEvaluateRequest(request: gatewayproto.EvaluateRequest): peer.Proposal {
    const proposalBytes = request.getProposedTransaction()?.getProposalBytes_asU8();
    expect(proposalBytes).toBeDefined(); // eslint-disable-line @typescript-eslint/no-non-null-assertion
    return peer.Proposal.deserializeBinary(proposalBytes!); // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeEndorseRequest(request: gatewayproto.EndorseRequest): peer.Proposal {
    const proposalBytes = request.getProposedTransaction()?.getProposalBytes_asU8();
    expect(proposalBytes).toBeDefined(); // eslint-disable-line @typescript-eslint/no-non-null-assertion
    return peer.Proposal.deserializeBinary(proposalBytes!); // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeChaincodeSpec(proposal: peer.Proposal): peer.ChaincodeSpec {
    const payload = peer.ChaincodeProposalPayload.deserializeBinary(proposal.getPayload_asU8());
    const invocationSpec = peer.ChaincodeInvocationSpec.deserializeBinary(payload.getInput_asU8());
    expect(invocationSpec.getChaincodeSpec()).toBeDefined();
    return invocationSpec.getChaincodeSpec()!; // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeArgsAsStrings(proposal: peer.Proposal): string[] {
    const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
    expect(chaincodeSpec.getInput()).toBeDefined();

    const args = chaincodeSpec.getInput()!.getArgsList_asU8(); // eslint-disable-line @typescript-eslint/no-non-null-assertion
    expect(args).toBeDefined();

    return args.map(arg => Buffer.from(arg).toString());
}

function assertDecodeHeader(proposal: peer.Proposal): common.Header {
    return common.Header.deserializeBinary(proposal.getHeader_asU8());
}

function assertDecodeSignatureHeader(proposal: peer.Proposal): common.SignatureHeader {
    const header = assertDecodeHeader(proposal);
    return common.SignatureHeader.deserializeBinary(header.getSignatureHeader_asU8());
}

function assertDecodeChannelHeader(proposal: peer.Proposal): common.ChannelHeader {
    const header = assertDecodeHeader(proposal);
    return common.ChannelHeader.deserializeBinary(header.getChannelHeader_asU8());
}

describe('Proposal', () => {
    const serviceError: ServiceError = Object.assign(new Error('ERROR_MESSAGE'), {
        code: status.UNAVAILABLE,
        details: 'DETAILS',
        metadata: new Metadata(),
    });

    let evaluateOptions: () => CallOptions;
    let endorseOptions: () => CallOptions;
    let client: MockGatewayGrpcClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(() => {
        const now = new Date();
        const evaluateCallOptions = {
            deadline: now.setHours(now.getHours() + 1),
        };
        evaluateOptions = () => evaluateCallOptions; // Return a specific object to test modification
        const endorseCallOptions = {
            deadline: now.setHours(now.getHours() + 1),
        };
        endorseOptions = () => endorseCallOptions; // Return a specific object to test modification

        client = new MockGatewayGrpcClient();
        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        };
        signer = jest.fn(undefined);
        signer.mockResolvedValue(Buffer.from('SIGNATURE'));
        hash = jest.fn(undefined);
        hash.mockReturnValue(Buffer.from('DIGEST'));

        gateway = internalConnect({
            identity,
            signer,
            hash,
            client,
            evaluateOptions,
            endorseOptions
        });
        network = gateway.getNetwork('CHANNEL_NAME');
        contract = network.getContract('CHAINCODE_NAME');
    });

    describe('evaluate', () => {
        const expectedResult = 'TX_RESULT';

        beforeEach(() => {
            const txResult = new peer.Response();
            txResult.setPayload(Buffer.from(expectedResult));

            const evaluateResult = new gatewayproto.EvaluateResponse();
            evaluateResult.setResult(txResult);

            client.mockEvaluateResponse(evaluateResult);
        });

        it('throws on evaluate error', async () => {
            client.mockEvaluateError(serviceError);

            await expect(contract.evaluateTransaction('TRANSACTION_NAME')).rejects.toThrow(serviceError.message);
        });

        it('returns result', async () => {
            const result = await contract.evaluateTransaction('TRANSACTION_NAME');

            const actual = Buffer.from(result).toString();
            expect(actual).toBe(expectedResult);
        });

        it('includes channel name in proposal', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.getEvaluateRequests()[0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const channelHeader = assertDecodeChannelHeader(proposal);
            expect(channelHeader.getChannelId()).toBe(network.getName());
        });

        it('includes chaincode name in proposal', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.getEvaluateRequests()[0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
            expect(chaincodeSpec.getChaincodeId()).toBeDefined();
            expect(chaincodeSpec.getChaincodeId()?.getName()).toBe(contract.getChaincodeName());
        });

        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_NAME');

            await contract.evaluateTransaction('MY_TRANSACTION');

            const evaluateRequest = client.getEvaluateRequests()[0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings[0]).toBe('MY_TRANSACTION');
        });

        it('includes transaction name in proposal for named smart contract', async () => {
            contract = network.getContract('CHAINCODE_NAME', 'MY_CONTRACT');

            await contract.evaluateTransaction('MY_TRANSACTION');

            const evaluateRequest = client.getEvaluateRequests()[0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings[0]).toBe('MY_CONTRACT:MY_TRANSACTION');
        });

        it('includes string arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];

            await contract.evaluateTransaction('TRANSACTION_NAME', ...expected);

            const evaluateRequest = client.getEvaluateRequests()[0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('includes bytes arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];
            const args = expected.map(arg => Buffer.from(arg));

            await contract.evaluateTransaction('TRANSACTION_NAME', ...args);

            const evaluateRequest = client.getEvaluateRequests()[0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('incldues bytes transient data in proposal', async () => {
            const transientData = {
                'uno': new Uint8Array(Buffer.from('one')),
                'dos': new Uint8Array(Buffer.from('two')),
            };
            await contract.evaluate('TRANSACTION_NAME', { transientData });

            const evaluateRequest = client.getEvaluateRequests()[0];
            const proposal_bytes = evaluateRequest.getProposedTransaction()?.getProposalBytes_asU8() ?? Buffer.from('');
            const proposal = peer.Proposal.deserializeBinary(proposal_bytes);
            const payload = peer.ChaincodeProposalPayload.deserializeBinary(proposal.getPayload_asU8());

            const actual = Object.fromEntries(payload.getTransientmapMap().getEntryList());

            expect(actual).toEqual(transientData);
        });

        it('incldues string transient data in proposal', async () => {
            const transientData = {
                'uno': 'one',
                'dos': 'two',
            };
            await contract.evaluate('TRANSACTION_NAME', { transientData });

            const evaluateRequest = client.getEvaluateRequests()[0];
            const proposal_bytes = evaluateRequest.getProposedTransaction()?.getProposalBytes_asU8() ?? Buffer.from('');
            const proposal = peer.Proposal.deserializeBinary(proposal_bytes);
            const payload = peer.ChaincodeProposalPayload.deserializeBinary(proposal.getPayload_asU8());

            const actual = Object.fromEntries(payload.getTransientmapMap().getEntryList());
            const expected: Record<string, Uint8Array> = {};
            Object.entries(transientData).forEach(([k, v]) => expected[k] = new Uint8Array(Buffer.from(v)));

            expect(actual).toEqual(expected);
        });

        it('sets endorsing orgs', async () => {
            await contract.evaluate('TRANSACTION_NAME', { endorsingOrganizations: ['org1']});

            const evaluateRequest = client.getEvaluateRequests()[0];
            const actualOrgs = evaluateRequest.getTargetOrganizationsList();
            expect(actualOrgs).toStrictEqual(['org1']);
        });

        it('uses signer', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.getEvaluateRequests()[0];
            const signature = Buffer.from(evaluateRequest.getProposedTransaction()?.getSignature_asU8() ?? '').toString();
            expect(signature).toBe('MY_SIGNATURE');
        });

        it('uses signer with newProposal', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const newProposal = gateway.newProposal(unsignedProposal.getBytes());
            await newProposal.evaluate();

            const evaluateRequest = client.getEvaluateRequests()[0];
            const signature = Buffer.from(evaluateRequest.getProposedTransaction()?.getSignature_asU8() ?? '').toString();
            expect(signature).toBe('MY_SIGNATURE');
        });

        it('uses hash', async () => {
            hash.mockReturnValue(Buffer.from('MY_DIGEST'));

            await contract.evaluateTransaction('TRANSACTION_NAME');

            expect(signer).toHaveBeenCalled();
            const digest = signer.mock.calls[0][0].toString();
            expect(digest).toBe('MY_DIGEST');
        });

        it('uses identity', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.getEvaluateRequests()[0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const signatureHeader = assertDecodeSignatureHeader(proposal);

            const expected = new msp.SerializedIdentity();
            expected.setMspid(identity.mspId);
            expected.setIdBytes(identity.credentials);

            expect(signatureHeader.getCreator()).toEqual(expected.serializeBinary());
        });

        it('includes channel name in request', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');

            const expected = network.getName();

            const evaluateRequest = client.getEvaluateRequests()[0];
            expect(evaluateRequest.getChannelId()).toBe(expected);

            const proposalProto = assertDecodeEvaluateRequest(evaluateRequest);
            const channelHeader = assertDecodeChannelHeader(proposalProto);
            expect(channelHeader.getChannelId()).toBe(expected);
        });

        it('includes transaction ID in request', async () => {
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await proposal.evaluate();

            const expected = proposal.getTransactionId();
            expect(expected).not.toHaveLength(0);

            const evaluateRequest = client.getEvaluateRequests()[0];
            expect(evaluateRequest.getTransactionId()).toBe(expected);

            const proposalProto = assertDecodeEvaluateRequest(evaluateRequest);
            const channelHeader = assertDecodeChannelHeader(proposalProto);
            expect(channelHeader.getTxId()).toBe(expected);
        });

        it('uses specified call options', async () => {
            const deadline = Date.now() + 1000;
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await proposal.evaluate({ deadline });

            const actual = client.getEvaluateOptions()[0];
            expect(actual.deadline).toBe(deadline);
        });

        it('uses default call options', async () => {
            await contract.evaluate('TRANSACTION_NAME');

            const actual = client.getEvaluateOptions()[0];
            expect(actual.deadline).toBe(evaluateOptions().deadline);
        });

        it('default call options are not modified', async () => {
            const expected = evaluateOptions().deadline;
            const deadline = Date.now() + 1000;
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await proposal.evaluate({ deadline });

            expect(evaluateOptions().deadline).toBe(expected);
        });
    });

    describe('endorse', () => {
        beforeEach(() => {
            const endorseResult = newEndorseResponse({
                result: Buffer.from('TX_RESULT'),
            });
            client.mockEndorseResponse(endorseResult);

            const commitResult = new gatewayproto.CommitStatusResponse();
            commitResult.setResult(peer.TxValidationCode.VALID);
            client.mockCommitStatusResponse(commitResult);
        });

        it('throws on endorse error', async () => {
            client.mockEndorseError(serviceError);
            const proposal = contract.newProposal('TRANSACTION_NAME');
            const transactionId = proposal.getTransactionId();

            const t = proposal.endorse();

            await expect(t).rejects.toThrow(EndorseError);
            await expect(t).rejects.toThrow(serviceError.message);
            await expect(t).rejects.toMatchObject({
                name: EndorseError.name,
                transactionId,
                code: serviceError.code,
                cause: serviceError,
            });
        });

        it('includes channel name in proposal', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.getEndorseRequests()[0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const channelHeader = assertDecodeChannelHeader(proposal);
            expect(channelHeader.getChannelId()).toBe(network.getName());
        });

        it('includes chaincode name in proposal', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.getEndorseRequests()[0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
            expect(chaincodeSpec.getChaincodeId()).toBeDefined();
            expect(chaincodeSpec.getChaincodeId()?.getName()).toBe(contract.getChaincodeName());
        });

        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_NAME');

            await contract.submitTransaction('MY_TRANSACTION');

            const endorseRequest = client.getEndorseRequests()[0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings[0]).toBe('MY_TRANSACTION');
        });

        it('includes transaction name in proposal for named smart contract', async () => {
            contract = network.getContract('CHAINCODE_NAME', 'MY_CONTRACT');

            await contract.submitTransaction('MY_TRANSACTION');

            const endorseRequest = client.getEndorseRequests()[0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings[0]).toBe('MY_CONTRACT:MY_TRANSACTION');
        });

        it('includes string arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];

            await contract.submitTransaction('TRANSACTION_NAME', ...expected);

            const endorseRequest = client.getEndorseRequests()[0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('includes bytes arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];
            const args = expected.map(arg => Buffer.from(arg));

            await contract.submitTransaction('TRANSACTION_NAME', ...args);

            const endorseRequest = client.getEndorseRequests()[0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('uses signer', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.getEndorseRequests()[0];
            const signature = Buffer.from(endorseRequest.getProposedTransaction()?.getSignature_asU8() ?? '').toString();
            expect(signature).toBe('MY_SIGNATURE');
        });

        it('uses signer with newProposal', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));
            const unsignedProposal = contract.newProposal('TRANSACTION_NAME');
            const newProposal = gateway.newProposal(unsignedProposal.getBytes());
            await newProposal.endorse();

            const endorseRequest = client.getEndorseRequests()[0];
            const signature = Buffer.from(endorseRequest.getProposedTransaction()?.getSignature_asU8() ?? '').toString();
            expect(signature).toBe('MY_SIGNATURE');
        });

        it('uses hash', async () => {
            hash.mockReturnValue(Buffer.from('MY_DIGEST'));

            await contract.submitTransaction('TRANSACTION_NAME');

            expect(signer).toHaveBeenCalled();
            const digest = signer.mock.calls[0][0].toString();
            expect(digest).toBe('MY_DIGEST');
        });

        it('uses identity', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.getEndorseRequests()[0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const signatureHeader = assertDecodeSignatureHeader(proposal);

            const expected = new msp.SerializedIdentity();
            expected.setMspid(identity.mspId);
            expected.setIdBytes(identity.credentials);

            expect(signatureHeader.getCreator_asU8()).toEqual(expected.serializeBinary());
        });

        it('includes channel name in request', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.getEndorseRequests()[0];
            expect(endorseRequest.getChannelId()).toBe(network.getName());
        });

        it('includes transaction ID in request', async () => {
            const proposal = contract.newProposal('TRANSACTION_NAME');
            const expected = proposal.getTransactionId();
            expect(expected).not.toHaveLength(0);

            const transaction = await proposal.endorse();
            expect(transaction.getTransactionId()).toBe(expected);

            const commit = await transaction.submit();
            expect(commit.getTransactionId()).toBe(expected);

            const endorseRequest = client.getEndorseRequests()[0];
            expect(endorseRequest.getTransactionId()).toBe(expected);

            const proposalProto = assertDecodeEndorseRequest(endorseRequest);
            const channelHeader = assertDecodeChannelHeader(proposalProto);
            expect(channelHeader.getTxId()).toBe(expected);
        });

        it('uses specified call options', async () => {
            const deadline = Date.now() + 1000;
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await proposal.endorse({ deadline });

            const actual = client.getEndorseOptions()[0];
            expect(actual.deadline).toBe(deadline);
        });

        it('uses default call options', async () => {
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await proposal.endorse();

            const actual = client.getEndorseOptions()[0];
            expect(actual.deadline).toBe(endorseOptions().deadline);
        });

        it('default call options are not modified', async () => {
            const expected = endorseOptions().deadline;
            const deadline = Date.now() + 1000;
            const proposal = contract.newProposal('TRANSACTION_NAME');

            await proposal.endorse({ deadline });

            expect(endorseOptions().deadline).toBe(expected);
        });
    });
});
