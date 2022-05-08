/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CloseableAsyncIterable } from '@hyperledger/fabric-gateway';
import { BaseEventListener } from './baseeventlistener';
import { EventListener } from './eventlistener';

export class CheckpointEventListener<T> implements EventListener<T> {
    readonly #checkpoint: (event: T) => Promise<void>;
    #eventListener: BaseEventListener<T>;

    constructor(events: CloseableAsyncIterable<T>, checkpoint: (event: T) => Promise<void>) {
        this.#eventListener = new BaseEventListener<T>(events);
        this.#checkpoint = checkpoint;
    }

    async next(): Promise<T> {
        const event = await this.#eventListener.next();
        await this.#checkpoint(event);
        return event;
    }

    close(): void {
        this.#eventListener.close();
    }
}
