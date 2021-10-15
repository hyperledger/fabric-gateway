/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CloseableAsyncIterable } from 'fabric-gateway';

export class EventListener<T> {
    #iterator: AsyncIterator<T>;
    #close: () => void;

    constructor(events: CloseableAsyncIterable<T>) {
        this.#iterator = events[Symbol.asyncIterator]();
        this.#close = events.close;
    }

    async next(): Promise<T> {
        const result = await this.#iterator.next();
        return result.value;
    }

    close() {
        this.#close();
    }
}
