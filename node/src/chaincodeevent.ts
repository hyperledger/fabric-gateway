/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEventsResponse } from './protos/gateway/gateway_pb';
import { ChaincodeEvent as ChaincodeEventProto } from './protos/peer/chaincode_event_pb';

export interface ChaincodeEvent {
    blockNumber: bigint;
    transactionId: string;
    chaincodeId: string;
    eventName: string;
    payload: Uint8Array;
}

export function newChaincodeEvents(responses: AsyncIterable<ChaincodeEventsResponse>): AsyncIterable<ChaincodeEvent> {
    return {
        async* [Symbol.asyncIterator]() { // eslint-disable-line @typescript-eslint/require-await
            for await (const response of responses) {
                const blockNumber = BigInt(response.getBlockNumber() ?? 0);
                const events = response.getEventsList() || [];
                for (const event of events) {
                    yield newChaincodeEvent(blockNumber, event);
                }
            }
        }
    };
}

function newChaincodeEvent(blockNumber: bigint, event: ChaincodeEventProto): ChaincodeEvent {
    return {
        blockNumber,
        chaincodeId: event.getChaincodeId() ?? '',
        eventName: event.getEventName() ?? '',
        transactionId: event.getTxId() ?? '',
        payload: event.getPayload_asU8() || new Uint8Array(),
    };
}
