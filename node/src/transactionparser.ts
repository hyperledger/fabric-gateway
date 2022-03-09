/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { inspect } from 'util';
import { assertDefined } from './gateway';
import { ChannelHeader, Envelope, Header, Payload } from './protos/common/common_pb';
import { ChaincodeAction } from './protos/peer/proposal_pb';
import { ProposalResponsePayload } from './protos/peer/proposal_response_pb';
import { ChaincodeActionPayload, Transaction, TransactionAction } from './protos/peer/transaction_pb';

export function parseTransactionEnvelope(envelope: Envelope): {
    channelName: string;
    result: Uint8Array;
} {
    const payload = Payload.deserializeBinary(envelope.getPayload_asU8());
    const header = assertDefined(payload.getHeader(), 'Missing header');

    return {
        channelName: parseChannelNameFromHeader(header),
        result: parseResultFromPayload(payload),
    };
}

function parseChannelNameFromHeader(header: Header): string {
    const channelHeader = ChannelHeader.deserializeBinary(header.getChannelHeader_asU8());
    return channelHeader.getChannelId();
}

function parseResultFromPayload(payload: Payload): Uint8Array {
    const transaction = Transaction.deserializeBinary(payload.getData_asU8());

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

function parseResultFromTransactionAction(transactionAction: TransactionAction): Uint8Array {
    const actionPayload = ChaincodeActionPayload.deserializeBinary(transactionAction.getPayload_asU8());
    const endorsedAction = assertDefined(actionPayload.getAction(), 'Missing endorsed action');
    const responsePayload = ProposalResponsePayload.deserializeBinary(endorsedAction.getProposalResponsePayload_asU8());
    const chaincodeAction = ChaincodeAction.deserializeBinary(responsePayload.getExtension_asU8());
    const chaincodeResponse = assertDefined(chaincodeAction.getResponse(), 'Missing chaincode response');
    return chaincodeResponse.getPayload_asU8();
}
