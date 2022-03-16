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
    #transactionID?: string;

    checkpoint(blockNumber: bigint, transactionID?: string): Promise<void> {
        if (blockNumber !== this.#blockNumber) {
            this.#blockNumber = blockNumber;
            this.#transactionID = undefined;
        }
        if (transactionID) {
            this.#transactionID = transactionID;
        }
        return Promise.resolve();
    }

    getBlockNumber(): bigint | undefined {
        return this.#blockNumber;
    }

    getTransactionId(): string | undefined {
        return this.#transactionID;
    }
}
