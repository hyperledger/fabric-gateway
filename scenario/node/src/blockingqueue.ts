/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { AsyncBarrier } from "./asyncbarrier";

export class BlockingQueue<T> {
    #elements: T[] = [];
    #barrier = new AsyncBarrier();

    put(element: T): void {
        this.#elements.push(element);
        this.#barrier.signal();
    }

    async get(): Promise<T> {
        let result: T | undefined;
        while (!(result = this.#elements.shift())) {
            await this.#barrier.wait();
        }

        return result;
    }
}
