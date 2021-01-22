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
import { protos, common, msp } from './protos/protos';

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

function assertDecodeProposal(proposedTransaction: protos.IProposedTransaction): protos.Proposal {
    expect(proposedTransaction.proposal).toBeDefined();
    expect(proposedTransaction.proposal!.proposal_bytes).toBeDefined(); // eslint-disable-line @typescript-eslint/no-non-null-assertion
    return protos.Proposal.decode(proposedTransaction.proposal!.proposal_bytes!); // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeChaincodeSpec(proposedTransaction: protos.IProposedTransaction): protos.IChaincodeSpec {
    const proposal = assertDecodeProposal(proposedTransaction);
    const payload = protos.ChaincodeProposalPayload.decode(proposal.payload);
    const invocationSpec = protos.ChaincodeInvocationSpec.decode(payload.input);
    expect(invocationSpec.chaincode_spec).toBeDefined();
    return invocationSpec.chaincode_spec!; // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeArgsAsStrings(proposedTransaction: protos.IProposedTransaction): string[] {
    const chaincodeSpec = assertDecodeChaincodeSpec(proposedTransaction);
    expect(chaincodeSpec.input).toBeDefined();

    const args = chaincodeSpec.input!.args; // eslint-disable-line @typescript-eslint/no-non-null-assertion
    expect(args).toBeDefined();

    return args!.map(arg => Buffer.from(arg).toString()); // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function assertDecodeHeader(proposedTransaction: protos.IProposedTransaction): common.Header {
    const proposal = assertDecodeProposal(proposedTransaction);
    return common.Header.decode(proposal.header);
}

function assertDecodeSignatureHeader(proposedTransaction: protos.IProposedTransaction): common.SignatureHeader {
    const header = assertDecodeHeader(proposedTransaction);
    return common.SignatureHeader.decode(header.signature_header);
}

function assertDecodeChannelHeader(proposedTransaction: protos.IProposedTransaction): common.ChannelHeader {
    const header = assertDecodeHeader(proposedTransaction);
    return common.ChannelHeader.decode(header.channel_header);
}

describe('Proposal', () => {
    let client: MockGatewayClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
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

        const options: InternalConnectOptions = {
            identity,
            signer,
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
                value: Buffer.from(expectedResult),
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
    
            const proposedTransaction = client.evaluate.mock.calls[0][0];
            const channelHeader = assertDecodeChannelHeader(proposedTransaction);
            expect(channelHeader.channel_id).toBe(network.getName());
        });

        it('includes chaincode ID in proposal', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');
    
            const chaincodeSpec = assertDecodeChaincodeSpec(client.evaluate.mock.calls[0][0]);
            expect(chaincodeSpec.chaincode_id).toBeDefined();
            expect(chaincodeSpec.chaincode_id!.name).toBe(contract.getChaincodeId()); // eslint-disable-line @typescript-eslint/no-non-null-assertion
        });
    
        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_ID');
    
            await contract.evaluateTransaction('MY_TRANSACTION');
    
            const argStrings = assertDecodeArgsAsStrings(client.evaluate.mock.calls[0][0]);
            expect(argStrings[0]).toBe('MY_TRANSACTION');
        });
    
        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_ID', 'MY_CONTRACT');
    
            await contract.evaluateTransaction('MY_TRANSACTION');
    
            const argStrings = assertDecodeArgsAsStrings(client.evaluate.mock.calls[0][0]);
            expect(argStrings[0]).toBe('MY_CONTRACT:MY_TRANSACTION');
        });
    
        it('includes string arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];
    
            await contract.evaluateTransaction('TRANSACTION_NAME', ...expected);
    
            const argStrings = assertDecodeArgsAsStrings(client.evaluate.mock.calls[0][0]);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('includes bytes arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];
            const args = expected.map(arg => Buffer.from(arg));
    
            await contract.evaluateTransaction('TRANSACTION_NAME', ...args);
    
            const argStrings = assertDecodeArgsAsStrings(client.evaluate.mock.calls[0][0]);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('uses signer', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.evaluateTransaction('TRANSACTION_NAME');
    
            const proposedTransaction = client.evaluate.mock.calls[0][0];
            const signature = Buffer.from(proposedTransaction.proposal?.signature ?? '').toString();
            expect(signature).toBe('MY_SIGNATURE');
        });

        it('uses identity', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.evaluateTransaction('TRANSACTION_NAME');
    
            const proposedTransaction = client.evaluate.mock.calls[0][0];
            const signatureHeader = assertDecodeSignatureHeader(proposedTransaction);

            const expected = msp.SerializedIdentity.encode({
                mspid: identity.mspId,
                id_bytes: identity.credentials
            }).finish();
            expect(signatureHeader.creator).toEqual(expected);
        });

        it('includes channel name in proposed transaction', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');
    
            const proposedTransaction = client.evaluate.mock.calls[0][0];
            expect(proposedTransaction.channelId).toBe(network.getName());
        });

        it('includes transaction ID in proposed transaction', async () => {
            await contract.evaluateTransaction('TRANSACTION_NAME');
    
            const proposedTransaction = client.evaluate.mock.calls[0][0];
            const expected = assertDecodeChannelHeader(proposedTransaction).tx_id;
            expect(proposedTransaction.txId).toBe(expected);
        });
    });

    describe('submit', () => {
        beforeEach(() => {
            client.endorse.mockResolvedValue({
                envelope: {
                    payload: Buffer.from('PAYLOAD'),
                },
                response: {
                    value: Buffer.from('TX_RESULT'),
                },
            });
        });

        it('throws on endorse error', async () => {
            client.endorse.mockRejectedValue(new Error('ERROR_MESSAGE'));
    
            await expect(contract.submitTransaction('TRANSACTION_NAME')).rejects.toThrow('ERROR_MESSAGE');
        });
    
        it('includes channel name in proposal', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');
    
            const proposedTransaction = client.endorse.mock.calls[0][0];
            const channelHeader = assertDecodeChannelHeader(proposedTransaction);
            expect(channelHeader.channel_id).toBe(network.getName());
        });

        it('includes chaincode ID in proposal', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');
    
            const chaincodeSpec = assertDecodeChaincodeSpec(client.endorse.mock.calls[0][0]);
            expect(chaincodeSpec.chaincode_id).toBeDefined();
            expect(chaincodeSpec.chaincode_id!.name).toBe(contract.getChaincodeId()); // eslint-disable-line @typescript-eslint/no-non-null-assertion
        });
    
        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_ID');
    
            await contract.submitTransaction('MY_TRANSACTION');
    
            const argStrings = assertDecodeArgsAsStrings(client.endorse.mock.calls[0][0]);
            expect(argStrings[0]).toBe('MY_TRANSACTION');
        });
    
        it('includes transaction name in proposal for default smart contract', async () => {
            contract = network.getContract('CHAINCODE_ID', 'MY_CONTRACT');
    
            await contract.submitTransaction('MY_TRANSACTION');
    
            const argStrings = assertDecodeArgsAsStrings(client.endorse.mock.calls[0][0]);
            expect(argStrings[0]).toBe('MY_CONTRACT:MY_TRANSACTION');
        });
    
        it('includes string arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];
    
            await contract.submitTransaction('TRANSACTION_NAME', ...expected);
    
            const argStrings = assertDecodeArgsAsStrings(client.endorse.mock.calls[0][0]);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('includes bytes arguments in proposal', async () => {
            const expected = ['one', 'two', 'three'];
            const args = expected.map(arg => Buffer.from(arg));
    
            await contract.submitTransaction('TRANSACTION_NAME', ...args);
    
            const argStrings = assertDecodeArgsAsStrings(client.endorse.mock.calls[0][0]);
            expect(argStrings.slice(1)).toStrictEqual(expected);
        });

        it('uses signer', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.submitTransaction('TRANSACTION_NAME');
    
            const proposedTransaction = client.endorse.mock.calls[0][0];
            const signature = Buffer.from(proposedTransaction.proposal?.signature ?? '').toString();
            expect(signature).toBe('MY_SIGNATURE');
        });

        it('uses identity', async () => {
            signer.mockResolvedValue(Buffer.from('MY_SIGNATURE'));

            await contract.submitTransaction('TRANSACTION_NAME');
    
            const proposedTransaction = client.endorse.mock.calls[0][0];
            const signatureHeader = assertDecodeSignatureHeader(proposedTransaction);

            const expected = msp.SerializedIdentity.encode({
                mspid: identity.mspId,
                id_bytes: identity.credentials
            }).finish();
            expect(signatureHeader.creator).toEqual(expected);
        });
    
        it('includes channel name in proposed transaction', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');
    
            const proposedTransaction = client.endorse.mock.calls[0][0];
            expect(proposedTransaction.channelId).toBe(network.getName());
        });

        it('includes transaction ID in proposed transaction', async () => {
            await contract.submitTransaction('TRANSACTION_NAME');
    
            const proposedTransaction = client.endorse.mock.calls[0][0];
            const expected = assertDecodeChannelHeader(proposedTransaction).tx_id;
            expect(proposedTransaction.txId).toBe(expected);
        });
    });
});
