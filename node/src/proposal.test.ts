/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { MockGatewayClient, newMockGatewayClient } from './client.test';
import { Contract } from './contract';
import { Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';
import { ChannelHeader, Envelope, Header, SignatureHeader } from './protos/common/common_pb';
import { CommitStatusResponse, EndorseRequest, EndorseResponse, EvaluateRequest, EvaluateResponse } from './protos/gateway/gateway_pb';
import { SerializedIdentity } from './protos/msp/identities_pb';
import { ChaincodeInvocationSpec, ChaincodeSpec } from './protos/peer/chaincode_pb';
import { ChaincodeProposalPayload, Proposal as ProposalProto } from './protos/peer/proposal_pb';
import { Response } from './protos/peer/proposal_response_pb';
import { TxValidationCode } from './protos/peer/transaction_pb';

function assertDecodeEvaluateRequest(request: EvaluateRequest): ProposalProto {
    const proposalBytes = request.getProposedTransaction()?.getProposalBytes_asU8();
    expect(proposalBytes).toBeDefined(); // eslint-disable-line @typescript-eslint/no-non-null-assertion
    return ProposalProto.deserializeBinary(proposalBytes!); // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeEndorseRequest(request: EndorseRequest): ProposalProto {
    const proposalBytes = request.getProposedTransaction()?.getProposalBytes_asU8();
    expect(proposalBytes).toBeDefined(); // eslint-disable-line @typescript-eslint/no-non-null-assertion
    return ProposalProto.deserializeBinary(proposalBytes!); // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeChaincodeSpec(proposal: ProposalProto): ChaincodeSpec {
    const payload = ChaincodeProposalPayload.deserializeBinary(proposal.getPayload_asU8());
    const invocationSpec = ChaincodeInvocationSpec.deserializeBinary(payload.getInput_asU8());
    expect(invocationSpec.getChaincodeSpec()).toBeDefined();
    return invocationSpec.getChaincodeSpec()!; // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeArgsAsStrings(proposal: ProposalProto): string[] {
    const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
    expect(chaincodeSpec.getInput()).toBeDefined();

    const args = chaincodeSpec.getInput()!.getArgsList_asU8(); // eslint-disable-line @typescript-eslint/no-non-null-assertion
    expect(args).toBeDefined();

    return args.map(arg => Buffer.from(arg).toString());
}

function assertDecodeHeader(proposal: ProposalProto): Header {
    return Header.deserializeBinary(proposal.getHeader_asU8());
}

function assertDecodeSignatureHeader(proposal: ProposalProto): SignatureHeader {
    const header = assertDecodeHeader(proposal);
    return SignatureHeader.deserializeBinary(header.getSignatureHeader_asU8());
}

function assertDecodeChannelHeader(proposal: ProposalProto): ChannelHeader {
    const header = assertDecodeHeader(proposal);
    return ChannelHeader.deserializeBinary(header.getChannelHeader_asU8());
}

describe('Proposal', () => {
    let client: MockGatewayClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(() => {
        client = newMockGatewayClient();
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
        contract = network.getContract('CHAINCODE_NAME');
    });

    describe('evaluate', () => {
        const expectedResult = 'TX_RESULT';

        beforeEach(() => {
            const txResult = new Response()
            txResult.setPayload(Buffer.from(expectedResult));
    
            const evaluateResult = new EvaluateResponse();
            evaluateResult.setResult(txResult)
    
            client.evaluate.mockResolvedValue(evaluateResult);
        });

        it('throws on evaluate error', async () => {
            client.evaluate.mockRejectedValue(new Error('ERROR_MESSAGE'));

            await expect(contract.evaluateTransaction('TRANSACTION_NAME')).rejects.toThrow('ERROR_MESSAGE');
        });

        it('returns result', async () => {
            const result = await contract.evaluateTransaction('TRANSACTION_NAME');

            const actual = Buffer.from(result).toString();
            expect(actual).toBe(expectedResult);
        });

        it('includes channel name in proposal', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const channelHeader = assertDecodeChannelHeader(proposal);
            expect(channelHeader.getChannelId()).toBe(network.getName());
        });

        it('includes chaincode name in proposal', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
            expect(chaincodeSpec.getChaincodeId()).toBeDefined();
            expect(chaincodeSpec.getChaincodeId()?.getName()).toBe(contract.getChaincodeName());
        });

        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_NAME');

            await contract.evaluateTransaction('MY_TRANSACTION');

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings[0]).toBe('MY_TRANSACTION');
        });

        it('includes transaction name in proposal for named smart contract', async () => {
            contract = network.getContract('CHAINCODE_NAME', 'MY_CONTRACT');

            await contract.evaluateTransaction('MY_TRANSACTION');

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings[0]).toBe('MY_CONTRACT:MY_TRANSACTION');
        });

        it('includes string arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];

            await contract.evaluateTransaction('TRANSACTION_NAME', ...expected);

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('includes bytes arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];
            const args = expected.map(arg => Buffer.from(arg));

            await contract.evaluateTransaction('TRANSACTION_NAME', ...args);

            const evaluateRequest = client.evaluate.mock.calls[0][0];
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

            const proposal_bytes = client.evaluate.mock.calls[0][0].getProposedTransaction()?.getProposalBytes_asU8() || Buffer.from('');
            const proposal = ProposalProto.deserializeBinary(proposal_bytes);
            const payload = ChaincodeProposalPayload.deserializeBinary(proposal.getPayload_asU8());

            const actual = Object.fromEntries(payload.getTransientmapMap().getEntryList());

            expect(actual).toEqual(transientData);
        });

        it('incldues string transient data in proposal', async () => {
            const transientData = {
                'uno': 'one',
                'dos': 'two',
            };
            await contract.evaluate('TRANSACTION_NAME', { transientData });

            const proposal_bytes = client.evaluate.mock.calls[0][0].getProposedTransaction()?.getProposalBytes_asU8() || Buffer.from('');
            const proposal = ProposalProto.deserializeBinary(proposal_bytes);
            const payload = ChaincodeProposalPayload.deserializeBinary(proposal.getPayload_asU8());

            const actual = Object.fromEntries(payload.getTransientmapMap().getEntryList());
            const expected: Record<string, Uint8Array> = {}
            Object.entries(transientData).forEach(([k, v]) => expected[k] = new Uint8Array(Buffer.from(v)));

            expect(actual).toEqual(expected)
        });

        it('sets endorsing orgs', async () => {
            await contract.evaluate('TRANSACTION_NAME', { endorsingOrganizations: ['org1']});
            const actualOrgs = client.evaluate.mock.calls[0][0].getTargetOrganizationsList();
            expect(actualOrgs).toStrictEqual(['org1']);
        });

        it('uses signer', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const signature = Buffer.from(evaluateRequest.getProposedTransaction()?.getSignature_asU8() || '').toString();
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

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const signatureHeader = assertDecodeSignatureHeader(proposal);

            const expected = new SerializedIdentity();
            expected.setMspid(identity.mspId);
            expected.setIdBytes(identity.credentials);

            expect(signatureHeader.getCreator()).toEqual(expected.serializeBinary());
        });

        it('includes channel name in request', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');

            const expected = network.getName();

            const evaluateRequest = client.evaluate.mock.calls[0][0];
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

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            expect(evaluateRequest.getTransactionId()).toBe(expected);

            const proposalProto = assertDecodeEvaluateRequest(evaluateRequest);
            const channelHeader = assertDecodeChannelHeader(proposalProto);
            expect(channelHeader.getTxId()).toBe(expected);
        });
    });

    describe('submit', () => {
        beforeEach(() => {
            const txResult = new Response()
            txResult.setPayload(Buffer.from('TX_RESULT'));
    
            const preparedTx = new Envelope();
            preparedTx.setPayload(Buffer.from('PAYLOAD'));
    
            const endorseResult = new EndorseResponse();
            endorseResult.setPreparedTransaction(preparedTx);
            endorseResult.setResult(txResult)
    
            client.endorse.mockResolvedValue(endorseResult);

            const commitResult = new CommitStatusResponse();
            commitResult.setResult(TxValidationCode.VALID);
    
            client.commitStatus.mockResolvedValue(commitResult);
        });

        it('throws on endorse error', async () => {
            client.endorse.mockRejectedValue(new Error('ERROR_MESSAGE'));

            await expect(contract.submitTransaction('TRANSACTION_NAME')).rejects.toThrow('ERROR_MESSAGE');
        });

        it('includes channel name in proposal', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.endorse.mock.calls[0][0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const channelHeader = assertDecodeChannelHeader(proposal);
            expect(channelHeader.getChannelId()).toBe(network.getName());
        });

        it('includes chaincode name in proposal', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.endorse.mock.calls[0][0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
            expect(chaincodeSpec.getChaincodeId()).toBeDefined();
            expect(chaincodeSpec.getChaincodeId()?.getName()).toBe(contract.getChaincodeName());
        });

        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_NAME');

            await contract.submitTransaction('MY_TRANSACTION');

            const endorseRequest = client.endorse.mock.calls[0][0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings[0]).toBe('MY_TRANSACTION');
        });

        it('includes transaction name in proposal for named smart contract', async () => {
            contract = network.getContract('CHAINCODE_NAME', 'MY_CONTRACT');

            await contract.submitTransaction('MY_TRANSACTION');

            const endorseRequest = client.endorse.mock.calls[0][0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings[0]).toBe('MY_CONTRACT:MY_TRANSACTION');
        });

        it('includes string arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];

            await contract.submitTransaction('TRANSACTION_NAME', ...expected);

            const endorseRequest = client.endorse.mock.calls[0][0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('includes bytes arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];
            const args = expected.map(arg => Buffer.from(arg));

            await contract.submitTransaction('TRANSACTION_NAME', ...args);

            const endorseRequest = client.endorse.mock.calls[0][0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('uses signer', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.endorse.mock.calls[0][0];
            const signature = Buffer.from(endorseRequest.getProposedTransaction()?.getSignature_asU8() || '').toString();
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

            const endorseRequest = client.endorse.mock.calls[0][0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const signatureHeader = assertDecodeSignatureHeader(proposal);

            const expected = new SerializedIdentity();
            expected.setMspid(identity.mspId);
            expected.setIdBytes(identity.credentials);

            expect(signatureHeader.getCreator_asU8()).toEqual(expected.serializeBinary());
        });

        it('includes channel name in request', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.endorse.mock.calls[0][0];
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

            const endorseRequest = client.endorse.mock.calls[0][0];
            expect(endorseRequest.getTransactionId()).toBe(expected);

            const proposalProto = assertDecodeEndorseRequest(endorseRequest);
            const channelHeader = assertDecodeChannelHeader(proposalProto);
            expect(channelHeader.getTxId()).toBe(expected);
        });
    });
});
