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
 */
export async function file(path: string): Promise<Checkpointer> {
    const filecheckpointer = new FileCheckPointer(path);
    await filecheckpointer.init();
    return filecheckpointer;
}

/**
 * Create a checkpointer that stores its state in memory only.
 */
export function inMemory(): Checkpointer {
    return new InMemoryCheckPointer();
}

/**
 * Wraps a chaincode event iterable with one that can be used to checkpoint processed events. Closing this async
 * iterable closes the wrapped iterable.
 * @param events - Chaincode events.
 * @param checkpointer - A checkpointer used to record processed events.
 */
export function checkpointChaincodeEvents(events: CloseableAsyncIterable<ChaincodeEvent>, checkpointer: Checkpointer): CheckpointAsyncIterable<ChaincodeEvent> {
    return newCheckpointAsyncIterable(
        events,
        event => {
            const checkpointBlockNumber = checkpointer.getBlockNumber();
            if (checkpointBlockNumber == undefined || event.blockNumber > checkpointBlockNumber) {
                return false;
            }
            return event.blockNumber < checkpointBlockNumber ||
                checkpointer.getTransactionIds().has(event.transactionId);
        },
        event => checkpointer.checkpoint(event.blockNumber, event.transactionId)
    );
}

function newCheckpointAsyncIterable<T>(
    events: CloseableAsyncIterable<T>,
    isCheckpointed: (event: T) => boolean,
    checkpoint: (event: T) => Promise<void>
): CheckpointAsyncIterable<T> {
    let lastEvent: T | undefined;

    return {
        async* [Symbol.asyncIterator]() { // eslint-disable-line @typescript-eslint/require-await
            for await (const event of events) {
                if (!isCheckpointed(event)) {
                    lastEvent = event;
                    yield event;
                }
            }
        },
        close: () => {
            events.close();
        },
        checkpoint: async () => {
            if (lastEvent) {
                await checkpoint(lastEvent);
            }
        }
    };
}
