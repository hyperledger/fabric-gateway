/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CloseableAsyncIterable } from '@hyperledger/fabric-gateway';
import { EventListener } from './eventlistener';

export class BaseEventListener<T> implements EventListener<T> {
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

    close(): void {
        this.#close();
    }
}
