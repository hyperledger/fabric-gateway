/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    ChannelHeader,
    Envelope,
    Header,
    Payload,
    SignatureHeader,
} from '@hyperledger/fabric-protos/lib/common/common_pb';
import { SerializedIdentity } from '@hyperledger/fabric-protos/lib/msp/identities_pb';
import { ChaincodeAction } from '@hyperledger/fabric-protos/lib/peer/proposal_pb';
import { ProposalResponsePayload } from '@hyperledger/fabric-protos/lib/peer/proposal_response_pb';
import {
    ChaincodeActionPayload,
    Transaction,
    TransactionAction,
} from '@hyperledger/fabric-protos/lib/peer/transaction_pb';
import { assertDefined } from './gateway';
import { Identity } from './identity';

export function parseTransactionEnvelope(envelope: Envelope): {
    identity: Identity;
    result: Uint8Array;
    transactionId: string;
} {
    const payload = Payload.deserializeBinary(envelope.getPayload_asU8());
    const header = assertDefined(payload.getHeader(), 'Missing header');
    const creator = parseCreatorFromHeader(header);

    return {
        identity: {
            mspId: creator.getMspid(),
            credentials: creator.getIdBytes_asU8(),
        },
        result: parseResultFromPayload(payload),
        transactionId: parseTransactionIdFromHeader(header),
    };
}

function parseTransactionIdFromHeader(header: Header): string {
    const channelHeader = ChannelHeader.deserializeBinary(header.getChannelHeader_asU8());
    return channelHeader.getTxId();
}

function parseCreatorFromHeader(header: Header): SerializedIdentity {
    const signatureHeader = SignatureHeader.deserializeBinary(header.getSignatureHeader_asU8());
    return SerializedIdentity.deserializeBinary(signatureHeader.getCreator_asU8());
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

    throw Object.assign(new Error(`No proposal response found: ${asString(errors)}`), {
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

function asString(value: unknown): string {
    if (typeof value === 'string') {
        return `'${value}'`;
    }

    if (Array.isArray(value)) {
        if (value.length === 0) {
            return '[]';
        }

        const contents = value.map(asString).join(', ');
        return `[ ${contents} ]`;
    }

    return String(value);
}
