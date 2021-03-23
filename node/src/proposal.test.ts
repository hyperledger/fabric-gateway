/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Contract } from './contract';
import { connect, Gateway, InternalConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';
import { GatewayClient } from './client';
import { common, gateway, msp, protos } from './protos/protos';

interface MockGatewayClient extends GatewayClient {
    endorse: jest.Mock<Promise<gateway.IEndorseResponse>, gateway.IEndorseRequest[]>,
    evaluate: jest.Mock<Promise<gateway.IEvaluateResponse>, gateway.IEvaluateRequest[]>,
    submit: jest.Mock<Promise<gateway.ISubmitResponse>, gateway.ISubmitRequest[]>,
}

function newMockGatewayClient(): MockGatewayClient {
    return {
        endorse: jest.fn(),
        evaluate: jest.fn(),
        submit: jest.fn(),
    };
}

function assertDecodeEvaluateRequest(request: gateway.IEvaluateRequest): protos.Proposal {
    expect(request.proposed_transaction).toBeDefined();
    expect(request.proposed_transaction!.proposal_bytes).toBeDefined(); // eslint-disable-line @typescript-eslint/no-non-null-assertion
    return protos.Proposal.decode(request.proposed_transaction!.proposal_bytes!); // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeEndorseRequest(request: gateway.IEndorseRequest): protos.Proposal {
    expect(request.proposed_transaction).toBeDefined();
    expect(request.proposed_transaction!.proposal_bytes).toBeDefined(); // eslint-disable-line @typescript-eslint/no-non-null-assertion
    return protos.Proposal.decode(request.proposed_transaction!.proposal_bytes!); // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeChaincodeSpec(proposal: protos.Proposal): protos.IChaincodeSpec {
    const payload = protos.ChaincodeProposalPayload.decode(proposal.payload);
    const invocationSpec = protos.ChaincodeInvocationSpec.decode(payload.input);
    expect(invocationSpec.chaincode_spec).toBeDefined();
    return invocationSpec.chaincode_spec!; // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeArgsAsStrings(proposal: protos.Proposal): string[] {
    const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
    expect(chaincodeSpec.input).toBeDefined();

    const args = chaincodeSpec.input!.args; // eslint-disable-line @typescript-eslint/no-non-null-assertion
    expect(args).toBeDefined();

    return args!.map(arg => Buffer.from(arg).toString()); // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeHeader(proposal: protos.Proposal): common.Header {
    return common.Header.decode(proposal.header);
}

function assertDecodeSignatureHeader(proposal: protos.Proposal): common.SignatureHeader {
    const header = assertDecodeHeader(proposal);
    return common.SignatureHeader.decode(header.signature_header);
}

function assertDecodeChannelHeader(proposal: protos.Proposal): common.ChannelHeader {
    const header = assertDecodeHeader(proposal);
    return common.ChannelHeader.decode(header.channel_header);
}

describe('Proposal', () => {
    let client: MockGatewayClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;
    let contract: Contract;

    beforeEach(async () => {
        client = newMockGatewayClient();
        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        }
        signer = jest.fn().mockReturnValue('SIGNATURE');
        hash = jest.fn().mockReturnValue('DIGEST');

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

    describe('evaluate', () => {
        const expectedResult = 'TX_RESULT';

        beforeEach(() => {
            client.evaluate.mockResolvedValue({
                result: {
                    payload: Buffer.from(expectedResult),
                },
            });
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
            expect(channelHeader.channel_id).toBe(network.getName());
        });

        it('includes chaincode ID in proposal', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
            expect(chaincodeSpec.chaincode_id).toBeDefined();
            expect(chaincodeSpec.chaincode_id!.name).toBe(contract.getChaincodeId()); // eslint-disable-line @typescript-eslint/no-non-null-assertion
        });

        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_ID');

            await contract.evaluateTransaction('MY_TRANSACTION');

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings[0]).toBe('MY_TRANSACTION');
        });

        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_ID', 'MY_CONTRACT');

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

        it('uses signer', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const signature = Buffer.from(evaluateRequest.proposed_transaction?.signature ?? '').toString();
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

            const expected = msp.SerializedIdentity.encode({
                mspid: identity.mspId,
                id_bytes: identity.credentials
            }).finish();
            expect(signatureHeader.creator).toEqual(expected);
        });

        it('includes channel name in request', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            expect(evaluateRequest.channel_id).toBe(network.getName());
        });

        it('includes transaction ID in request', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');

            const evaluateRequest = client.evaluate.mock.calls[0][0];
            const proposal = assertDecodeEvaluateRequest(evaluateRequest);
            const expected = assertDecodeChannelHeader(proposal).tx_id;
            expect(evaluateRequest.transaction_id).toBe(expected);
        });
    });

    describe('submit', () => {
        beforeEach(() => {
            client.endorse.mockResolvedValue({
                prepared_transaction: {
                    payload: Buffer.from('PAYLOAD'),
                },
                result: {
                    payload: Buffer.from('TX_RESULT'),
                },
            });
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
            expect(channelHeader.channel_id).toBe(network.getName());
        });

        it('includes chaincode ID in proposal', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.endorse.mock.calls[0][0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const chaincodeSpec = assertDecodeChaincodeSpec(proposal);
            expect(chaincodeSpec.chaincode_id).toBeDefined();
            expect(chaincodeSpec.chaincode_id!.name).toBe(contract.getChaincodeId()); // eslint-disable-line @typescript-eslint/no-non-null-assertion
        });

        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_ID');

            await contract.submitTransaction('MY_TRANSACTION');

            const endorseRequest = client.endorse.mock.calls[0][0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const argStrings = assertDecodeArgsAsStrings(proposal);
            expect(argStrings[0]).toBe('MY_TRANSACTION');
        });

        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_ID', 'MY_CONTRACT');

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
            const signature = Buffer.from(endorseRequest.proposed_transaction?.signature ?? '').toString();
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

            const expected = msp.SerializedIdentity.encode({
                mspid: identity.mspId,
                id_bytes: identity.credentials
            }).finish();
            expect(signatureHeader.creator).toEqual(expected);
        });

        it('includes channel name in request', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.endorse.mock.calls[0][0];
            expect(endorseRequest.channel_id).toBe(network.getName());
        });

        it('includes transaction ID in request', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');

            const endorseRequest = client.endorse.mock.calls[0][0];
            const proposal = assertDecodeEndorseRequest(endorseRequest);
            const expected = assertDecodeChannelHeader(proposal).tx_id;
            expect(endorseRequest.transaction_id).toBe(expected);
        });
    });
});
