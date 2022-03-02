/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Checkpointer } from './checkpointer';

/**
 * In-memory checkpointer class used to persist checkpointer state in memory.
 */

export class InMemoryCheckPointer implements Checkpointer {

    #blockNumber?: bigint;
    #transactionIDs: Set<string> = new Set();

    checkpoint(blockNumber: bigint, transactionID?: string): Promise<void> {
        if (blockNumber !== this.#blockNumber) {
            this.#blockNumber = blockNumber;
            this.#transactionIDs.clear();
        }
        if (transactionID) {
            this.#transactionIDs.add(transactionID);
        }
        return Promise.resolve();
    }

    getBlockNumber(): bigint | undefined{
        return this.#blockNumber;
    }

    getTransactionIds(): Set<string> {
        return this.#transactionIDs;
    }
}