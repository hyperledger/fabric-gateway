/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { common, gateway as gatewayproto, msp, peer } from '@hyperledger/fabric-protos';
import { Contract, Gateway, Identity, Network, Proposal, connect } from '../src';
import { assertDefined } from '../src/gateway';

const utf8Encoder = new TextEncoder();
const utf8Decoder = new TextDecoder();

function decodeProposedTransaction(proposal: Proposal): gatewayproto.ProposedTransaction {
    return gatewayproto.ProposedTransaction.deserializeBinary(proposal.getBytes());
}

function assertDecodeProposal(proposal: Proposal): peer.Proposal {
    const proposedTransaction = decodeProposedTransaction(proposal);
    let proposalBytes = proposedTransaction.getProposal()?.getProposalBytes_asU8();
    proposalBytes = assertDefined(proposalBytes, 'proposalBytes is undefined');
    return peer.Proposal.deserializeBinary(proposalBytes);
}

function assertDecodeChaincodeSpec(proposal: peer.Proposal): peer.ChaincodeSpec {
    const payload = peer.ChaincodeProposalPayload.deserializeBinary(proposal.getPayload_asU8());
    const invocationSpec = peer.ChaincodeInvocationSpec.deserializeBinary(payload.getInput_asU8());
    const chaincodeSpec = invocationSpec.getChaincodeSpec();
    return assertDefined(chaincodeSpec, 'chaincodeSpec is undefined');
}

function assertDecodeArgsAsStrings(proposal: peer.Proposal): string[] {
    const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
    const input = assertDefined(chaincodeSpec.getInput(), 'input is undefined');
    const args = input.getArgsList_asU8();
    return args.map((arg) => utf8Decoder.decode(arg));
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
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(() => {
        identity = {
            mspId: 'MSP_ID',
            credentials: utf8Encoder.encode('CERTIFICATE'),
        };
        signer = jest.fn(undefined);
        signer.mockResolvedValue(utf8Encoder.encode('SIGNATURE'));

        gateway = connect({
            identity,
            signer,
        });
        network = gateway.getNetwork('CHANNEL_NAME');
        contract = network.getContract('CHAINCODE_NAME');
    });

    it('includes channel name', async () => {
        const result = await contract.newProposal('TRANSACTION_NAME');

        const proposal = assertDecodeProposal(result);
        const channelHeader = assertDecodeChannelHeader(proposal);
        expect(channelHeader.getChannelId()).toBe(network.getName());
    });

    it('includes chaincode name', async () => {
        const result = await contract.newProposal('TRANSACTION_NAME');

        const proposal = assertDecodeProposal(result);
        const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
        expect(chaincodeSpec.getChaincodeId()).toBeDefined();
        expect(chaincodeSpec.getChaincodeId()?.getName()).toBe(contract.getChaincodeName());
    });

    it('includes transaction name for default smart contract', async () => {
        const result = await contract.newProposal('MY_TRANSACTION');

        const proposal = assertDecodeProposal(result);
        const argStrings = assertDecodeArgsAsStrings(proposal);
        expect(argStrings[0]).toBe('MY_TRANSACTION');
    });

    it('includes transaction name for named smart contract', async () => {
        contract = network.getContract('CHAINCODE_NAME', 'MY_CONTRACT');
        const result = await contract.newProposal('MY_TRANSACTION');

        const proposal = assertDecodeProposal(result);
        const argStrings = assertDecodeArgsAsStrings(proposal);
        expect(argStrings[0]).toBe('MY_CONTRACT:MY_TRANSACTION');
    });

    it('includes string arguments', async () => {
        const expected = ['one', 'two', 'three'];
        const result = await contract.newProposal('TRANSACTION_NAME', {
            arguments: expected,
        });

        const proposal = assertDecodeProposal(result);
        const argStrings = assertDecodeArgsAsStrings(proposal);
        expect(argStrings.slice(1)).toStrictEqual(expected);
    });

    it('includes bytes arguments', async () => {
        const expected = ['one', 'two', 'three'];
        const args = expected.map((arg) => utf8Encoder.encode(arg));

        const result = await contract.newProposal('TRANSACTION_NAME', {
            arguments: args,
        });

        const proposal = assertDecodeProposal(result);
        const argStrings = assertDecodeArgsAsStrings(proposal);
        expect(argStrings.slice(1)).toStrictEqual(expected);
    });

    it('incldues bytes transient data', async () => {
        const transientData = {
            uno: new Uint8Array(utf8Encoder.encode('one')),
            dos: new Uint8Array(utf8Encoder.encode('two')),
        };
        const result = await contract.newProposal('TRANSACTION_NAME', {
            transientData,
        });

        const proposal = assertDecodeProposal(result);
        const payload = peer.ChaincodeProposalPayload.deserializeBinary(proposal.getPayload_asU8());

        const actual = Object.fromEntries(payload.getTransientmapMap().getEntryList());

        expect(actual).toEqual(transientData);
    });

    it('incldues string transient data', async () => {
        const transientData = {
            uno: 'one',
            dos: 'two',
        };
        const result = await contract.newProposal('TRANSACTION_NAME', {
            transientData,
        });

        const proposal = assertDecodeProposal(result);
        const payload = peer.ChaincodeProposalPayload.deserializeBinary(proposal.getPayload_asU8());

        const actual = Object.fromEntries(payload.getTransientmapMap().getEntryList());
        const expected: Record<string, Uint8Array> = {};
        Object.entries(transientData).forEach(([k, v]) => (expected[k] = utf8Encoder.encode(v)));

        expect(actual).toEqual(expected);
    });

    it('sets endorsing orgs', async () => {
        const result = await contract.newProposal('TRANSACTION_NAME', {
            endorsingOrganizations: ['org1'],
        });

        const proposedTransaction = decodeProposedTransaction(result);
        const actualOrgs = proposedTransaction.getEndorsingOrganizationsList();
        expect(actualOrgs).toStrictEqual(['org1']);
    });

    it('uses signer', async () => {
        signer.mockResolvedValue(utf8Encoder.encode('MY_SIGNATURE'));

        const result = await contract.newProposal('TRANSACTION_NAME');

        const proposedTransaction = decodeProposedTransaction(result);
        const signature = proposedTransaction.getProposal()?.getSignature_asU8() ?? new Uint8Array();
        expect(utf8Decoder.decode(signature)).toBe('MY_SIGNATURE');
    });

    it('uses identity', async () => {
        const result = await contract.newProposal('TRANSACTION_NAME');

        const proposal = assertDecodeProposal(result);
        const signatureHeader = assertDecodeSignatureHeader(proposal);

        const expected = new msp.SerializedIdentity();
        expected.setMspid(identity.mspId);
        expected.setIdBytes(identity.credentials);

        expect(signatureHeader.getCreator()).toEqual(expected.serializeBinary());
    });

    it('includes transaction ID', async () => {
        const result = await contract.newProposal('TRANSACTION_NAME');

        const proposal = assertDecodeProposal(result);
        const channelHeader = assertDecodeChannelHeader(proposal);
        const proposalTransactionId = channelHeader.getTxId();

        expect(proposalTransactionId).toHaveLength(64); // SHA-256 hash should be 32 bytes, which is 64 hex characters
        expect(result.getTransactionId()).toStrictEqual(proposalTransactionId);
    });
});
