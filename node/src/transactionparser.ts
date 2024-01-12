/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { common, peer } from '@hyperledger/fabric-protos';
import { inspect } from 'util';
import { assertDefined } from './gateway';

export function parseTransactionEnvelope(envelope: common.Envelope): {
    channelName: string;
    result: Uint8Array;
} {
    const payload = common.Payload.deserializeBinary(envelope.getPayload_asU8());
    const header = assertDefined(payload.getHeader(), 'Missing header');

    return {
        channelName: parseChannelNameFromHeader(header),
        result: parseResultFromPayload(payload),
    };
}

function parseChannelNameFromHeader(header: common.Header): string {
    const channelHeader = common.ChannelHeader.deserializeBinary(header.getChannelHeader_asU8());
    return channelHeader.getChannelId();
}

function parseResultFromPayload(payload: common.Payload): Uint8Array {
    const transaction = peer.Transaction.deserializeBinary(payload.getData_asU8());

    const errors: unknown[] = [];

    for (const transactionAction of transaction.getActionsList()) {
        try {
            return parseResultFromTransactionAction(transactionAction);
        } catch (err) {
            errors.push(err);
        }
    }

    throw Object.assign(new Error(`No proposal response found: ${inspect(errors)}`), {
        suppressed: errors,
    });
}

function parseResultFromTransactionAction(transactionAction: peer.TransactionAction): Uint8Array {
    const actionPayload = peer.ChaincodeActionPayload.deserializeBinary(transactionAction.getPayload_asU8());
    const endorsedAction = assertDefined(actionPayload.getAction(), 'Missing endorsed action');
    const responsePayload = peer.ProposalResponsePayload.deserializeBinary(
        endorsedAction.getProposalResponsePayload_asU8(),
    );
    const chaincodeAction = peer.ChaincodeAction.deserializeBinary(responsePayload.getExtension_asU8());
    const chaincodeResponse = assertDefined(chaincodeAction.getResponse(), 'Missing chaincode response');
    return chaincodeResponse.getPayload_asU8();
}
