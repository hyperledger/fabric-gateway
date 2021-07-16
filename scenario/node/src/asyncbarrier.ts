/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { EventEmitter } from 'events';

const eventName = 'event';

export class AsyncBarrier {
    #emitter = new EventEmitter();

    async wait(): Promise<void> {
        await new Promise(resolve => this.#emitter.once(eventName, resolve));
    }

    signal(): void {
        this.#emitter.emit(eventName);
    }
}
