/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { common, gateway as gatewayproto, peer } from '@hyperledger/fabric-protos';
import { Gateway, connect, Transaction } from '.';
import { assertDefined } from './gateway';
import { SigningIdentity } from './signingidentity';
import { TransactionContext } from './transactioncontext';

const utf8Encoder = new TextEncoder();
const utf8Decoder = new TextDecoder();

function newPreparedTransaction(options: {
    context: TransactionContext;
    result?: Uint8Array;
}): gatewayproto.PreparedTransaction {
    const chaincodeResponse = new peer.Response();
    chaincodeResponse.setPayload(options.result ?? new Uint8Array());

    const chaincodeAction = new peer.ChaincodeAction();
    chaincodeAction.setResponse(chaincodeResponse);

    const responsePayload = new peer.ProposalResponsePayload();
    responsePayload.setExtension$(chaincodeAction.serializeBinary());

    const endorsedAction = new peer.ChaincodeEndorsedAction();
    endorsedAction.setProposalResponsePayload(responsePayload.serializeBinary());

    const actionPayload = new peer.ChaincodeActionPayload();
    actionPayload.setAction(endorsedAction);

    const transactionAction = new peer.TransactionAction();
    transactionAction.setPayload(actionPayload.serializeBinary());

    const transaction = new peer.Transaction();
    transaction.setActionsList([transactionAction]);

    const channelHeader = new common.ChannelHeader();
    channelHeader.setTxId(options.context.getTransactionId());

    const header = new common.Header();
    header.setSignatureHeader(options.context.getSignatureHeader().serializeBinary());
    header.setChannelHeader(channelHeader.serializeBinary());

    const payload = new common.Payload();
    payload.setData(transaction.serializeBinary());
    payload.setHeader(header);

    const envelope = new common.Envelope();
    envelope.setPayload(payload.serializeBinary());

    const result = new gatewayproto.PreparedTransaction();
    result.setEnvelope(envelope);
    result.setTransactionId(options.context.getTransactionId());

    return result;
}

function assertDecodeSignature(transaction: Transaction): Uint8Array {
    const preparedTransaction = decodePreparedTransaction(transaction);
    const envelope = assertDefined(preparedTransaction.getEnvelope(), 'envelope is undefined');
    return envelope.getSignature_asU8();
}

function decodePreparedTransaction(transaction: Transaction): gatewayproto.PreparedTransaction {
    return gatewayproto.PreparedTransaction.deserializeBinary(transaction.getBytes());
}

describe('Transaction', () => {
    let signingIdentity: SigningIdentity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let gateway: Gateway;
    let context: TransactionContext;

    beforeEach(async () => {
        signer = jest.fn(undefined);
        signer.mockResolvedValue(utf8Encoder.encode('SIGNATURE'));

        signingIdentity = new SigningIdentity({
            identity: {
                mspId: 'MSP_ID',
                credentials: utf8Encoder.encode('CERTIFICATE'),
            },
            signer,
        });

        gateway = connect({
            identity: signingIdentity.getIdentity(),
            signer,
        });

        context = await TransactionContext.newInstance(signingIdentity);
    });

    it('newTransaction throws on identity MSP ID mismatch', async () => {
        const identity = signingIdentity.getIdentity();
        identity.mspId = 'WRONG_MSP_ID';
        context = await TransactionContext.newInstance(new SigningIdentity({ identity, signer }));
        const preparedTransaction = newPreparedTransaction({ context });

        const result = gateway.newTransaction(preparedTransaction.serializeBinary());

        await expect(result).rejects.toBeDefined();
    });

    it('newTransaction throws on identity credentials mismatch', async () => {
        const identity = signingIdentity.getIdentity();
        identity.credentials = utf8Encoder.encode('WRONG_CREDENTIALS');
        context = await TransactionContext.newInstance(new SigningIdentity({ identity, signer }));
        const preparedTransaction = newPreparedTransaction({ context });

        const result = gateway.newTransaction(preparedTransaction.serializeBinary());

        await expect(result).rejects.toBeDefined();
    });

    it('uses signer', async () => {
        signer.mockResolvedValue(utf8Encoder.encode('MY_SIGNATURE'));
        const preparedTransaction = newPreparedTransaction({ context });

        const result = await gateway.newTransaction(preparedTransaction.serializeBinary());

        const signature = utf8Decoder.decode(assertDecodeSignature(result));
        expect(signature).toBe('MY_SIGNATURE');
    });

    it('has correct transaction ID', async () => {
        const preparedTransaction = newPreparedTransaction({ context });

        const result = await gateway.newTransaction(preparedTransaction.serializeBinary());

        expect(result.getTransactionId()).toBe(context.getTransactionId());
    });

    it('uses transaction ID from signed content', async () => {
        const preparedTransaction = newPreparedTransaction({ context });
        preparedTransaction.setTransactionId('WRONG_TRANSACTION_ID');

        const result = await gateway.newTransaction(preparedTransaction.serializeBinary());

        expect(result.getTransactionId()).toBe(context.getTransactionId());
    });

    it('has correct result', async () => {
        const preparedTransaction = newPreparedTransaction({
            context,
            result: utf8Encoder.encode('MY_RESULT'),
        });

        const result = await gateway.newTransaction(preparedTransaction.serializeBinary());

        const actual = utf8Decoder.decode(result.getResult());
        expect(actual).toEqual('MY_RESULT');
    });
});
