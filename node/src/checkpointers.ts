/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEvent } from './chaincodeevent';
import { Checkpointer, CheckpointAsyncIterable } from './checkpointer';
import { CloseableAsyncIterable } from './client';
import { FileCheckPointer } from './filecheckpointer';
import { InMemoryCheckPointer } from './inmemorycheckpointer';

/**
 * Create a checkpointer that uses the specified file to store persistent state.
 * @param path - Path to a file holding persistent checkpoint state.
 * @returns Promise<FileCheckPointer> A file checkpointer.
 */
export async function file(path: string):Promise<Checkpointer>{
    const filecheckpointer = new FileCheckPointer(path);
    await filecheckpointer.init();
    return filecheckpointer;
}

/**
 *
 * @returns InMemoryCheckPointer An in-memory checkpointer.
 */
export function inMemory():Checkpointer{
    return new InMemoryCheckPointer();
}

/**
 *
 * @param events - An async iterator of events
 * @param checkpointer - A checkpointer instance to checkpoint the processed events.
 * @returns
 */
export function checkpointChaincodeEvents(events: CloseableAsyncIterable<ChaincodeEvent>, checkpointer: Checkpointer): CheckpointAsyncIterable<ChaincodeEvent> {
    function isCheckpointed(event: ChaincodeEvent) {
        const checkpointBlockNumber = checkpointer.getBlockNumber();
        if (checkpointBlockNumber && checkpointBlockNumber > event.blockNumber) {
            return true;
        }
        return checkpointer.getTransactionIds().has(event.transactionId);
    }
    return newCheckpointedIterable(events, isCheckpointed, event => checkpointer.checkpoint(event.blockNumber, event.transactionId));
}

function newCheckpointedIterable<T>(events: CloseableAsyncIterable<T>, isCheckpointed: (event: T) => boolean, checkpoint: (event: T) => Promise<void>): CheckpointAsyncIterable<T> {
    let lastEvent: T | undefined;

    return {
        async* [Symbol.asyncIterator]() { // eslint-disable-line @typescript-eslint/require-await
            for await (const event of events) {
                if (!isCheckpointed(event)) {
                    lastEvent = event
                    yield event;
                }
            }
        },
        close: () => {
            events.close();
        },
        //checkpoints the last yielded event
        checkpoint: async () => {
            if (lastEvent) {
                await checkpoint(lastEvent);
            }
        }
    };
}