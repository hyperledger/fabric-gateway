/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEvent, Checkpointer, CloseableAsyncIterable } from '@hyperledger/fabric-gateway';


export class CheckpointEventListener<T> {
    #iterator?: AsyncIterator<T>;
    #close?: () => void;
    #checkpointer: Checkpointer;

    constructor(checkpointer: Checkpointer) {
        this.#checkpointer = checkpointer;
    }

    setEvents(events: CloseableAsyncIterable<T>): void {
        this.#iterator = events[Symbol.asyncIterator]();
        this.#close = () => events.close();
    }

    getCheckpointer(): Checkpointer {
        return this.#checkpointer;
    }

    async next(): Promise<T> {
        const result = await this.#iterator?.next();
        const event = result?.value as T;
        if (event) {
            await this.#checkpointer.checkpointChaincodeEvent(event as unknown as ChaincodeEvent);
        }
        return event;
    }

    close(): void {
        this.#close? this.#close() : null;
    }
}
