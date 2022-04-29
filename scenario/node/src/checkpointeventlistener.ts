/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CloseableAsyncIterable } from '@hyperledger/fabric-gateway';
import { EventListener } from './eventlistener';

export class CheckpointEventListener<T> extends EventListener<T> {
    readonly #checkpoint: (event: T) => Promise<void>;

    constructor(events: CloseableAsyncIterable<T>, checkpoint: (event: T) => Promise<void>) {
        super(events);
        this.#checkpoint = checkpoint;
    }

    async next(): Promise<T> {
        const event = await super.next();
        await this.#checkpoint(event);
        return event;
    }
}
