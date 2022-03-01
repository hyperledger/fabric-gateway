/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEventsOptions } from './chaincodeeventsbuilder';
import { CheckPointer } from './checkpointer';
import { CloseableAsyncIterable } from './client';
import { ChaincodeEventsResponse } from './protos/gateway/gateway_pb';
import { ChaincodeEvent as ChaincodeEventProto } from './protos/peer/chaincode_event_pb';

/**
 * Chaincode event emitted by a transaction function.
 */
export interface ChaincodeEvent {
    /**
     * Block number that included this chaincode event.
     */
    blockNumber: bigint;

    /**
     * Transaction that emitted this chaincode event.
     */
    transactionId: string;

    /**
     * Chaincode that emitted this event.
     */
    chaincodeName: string;

    /**
     * Name of the emitted event.
     */
    eventName: string;

    /**
     * Application defined payload data associated with this event.
     */
    payload: Uint8Array;
}

export function newChaincodeEvents(responses: CloseableAsyncIterable<ChaincodeEventsResponse>, eventOptions?:ChaincodeEventsOptions): CloseableAsyncIterable<ChaincodeEvent> {

    const checkPointer: CheckPointer|undefined = eventOptions?.checkPointer

    return {

        async* [Symbol.asyncIterator]() { // eslint-disable-line @typescript-eslint/require-await

            for await (const response of responses) {
                const blockNumber = BigInt(response.getBlockNumber() ?? 0); //add explanation
                const blockCheckPointed:boolean = await isCheckPointed(checkPointer,blockNumber);

                if(!blockCheckPointed){
                    const events = response.getEventsList() || [];
                for (const event of events) {
                    const transactionID = event.getTxId();
                    const eventCheckPointed:boolean = await isCheckPointed(checkPointer,undefined,transactionID);
                    if(!eventCheckPointed){
                        yield newChaincodeEvent(blockNumber, event);
                    }
                }
            }
            }
        },
        close: () => {
            responses.close();
        },
    };
}

export async function isCheckPointed(checkPointer?:CheckPointer,blocNumber?:bigint,transactionID?:string): Promise<boolean>{
    if(checkPointer){

    if(blocNumber){
        const blockNumberCheckPointed = await checkPointer.getBlockNumber();
        if(blockNumberCheckPointed && blockNumberCheckPointed >blocNumber){
            return true
        }
    }
    if(transactionID){
        const transactionIDs = await checkPointer.getTransactionIds();
        if(!transactionIDs.includes(transactionID)){
            return true
        }
    }

    }
    return false
}
function newChaincodeEvent(blockNumber: bigint, event: ChaincodeEventProto): ChaincodeEvent {
    return {
        blockNumber,
        chaincodeName: event.getChaincodeId() ?? '',
        eventName: event.getEventName() ?? '',
        transactionId: event.getTxId() ?? '',
        payload: event.getPayload_asU8() || new Uint8Array(),
    };
}
