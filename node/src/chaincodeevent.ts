/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as Long from 'long';
import { gateway, protos } from './protos/protos';

export interface ChaincodeEvent {
    blockNumber: Long;
    transactionId: string;
    chaincodeId: string;
    eventName: string;
    payload: Uint8Array;
}

export function newChaincodeEvents(responses: AsyncIterable<gateway.IChaincodeEventsResponse>): AsyncIterable<ChaincodeEvent> {
    return {
        async* [Symbol.asyncIterator]() { // eslint-disable-line @typescript-eslint/require-await
            for await (const response of responses) {
                const blockNumber = Long.fromValue(response.block_number ?? 0);
                const events = response.events || [];
                for (const event of events) {
                    yield newChaincodeEvent(blockNumber, event);
                }
            }
        }
    };
}

function newChaincodeEvent(blockNumber: Long, event: protos.IChaincodeEvent): ChaincodeEvent {
    return {
        blockNumber,
        chaincodeId: event.chaincode_id ?? '',
        eventName: event.event_name ?? '',
        transactionId: event.tx_id ?? '',
        payload: event.payload ?? new Uint8Array(),
    };
}
