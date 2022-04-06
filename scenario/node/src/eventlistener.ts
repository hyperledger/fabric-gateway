/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Checkpointer, CloseableAsyncIterable } from '@hyperledger/fabric-gateway';

export class EventListener<T> {
    #iterator: AsyncIterator<T>;
    #close: () => void;

    constructor(events: CloseableAsyncIterable<T>) {
        this.#iterator = events[Symbol.asyncIterator]();
        this.#close = () => events.close();
    }

    async next(): Promise<T> {
        const result = await this.#iterator.next();
        return result.value as T;
    }

    async checkpointBlock(checkpointer: Checkpointer, blockNumber: bigint): Promise<void> {
        await checkpointer.checkpointBlock(blockNumber);
    }

    close(): void {
        this.#close();
    }
}
