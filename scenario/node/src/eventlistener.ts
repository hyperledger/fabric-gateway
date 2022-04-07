/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEvent, Checkpointer, CloseableAsyncIterable } from '@hyperledger/fabric-gateway';

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

    async checkpointChaincodeEvent(checkpointer: Checkpointer, event:ChaincodeEvent): Promise<void> {
        await checkpointer.checkpointChaincodeEvent(event);
    }

    close(): void {
        this.#close();
    }
}
