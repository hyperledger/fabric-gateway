/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Checkpointer } from './checkpointer';
import { ChaincodeEvent } from './chaincodeevent';

/**
 * In-memory checkpointer class used to persist checkpointer state in memory.
 */

export class InMemoryCheckPointer implements Checkpointer {
    #blockNumber?: bigint;
    #transactionId?: string;

    checkpointBlock(blockNumber: bigint): Promise<void> {
        this.#blockNumber = blockNumber + BigInt(1);
        this.#transactionId = undefined;
        return Promise.resolve();
    }

    checkpointTransaction(blockNumber: bigint, transactionId: string): Promise<void> {
        this.#blockNumber = blockNumber;
        this.#transactionId = transactionId;
        return Promise.resolve();
    }

    async checkpointChaincodeEvent(event: ChaincodeEvent): Promise<void> {
        await this.checkpointTransaction(event.blockNumber, event.transactionId);
    }

    getBlockNumber(): bigint | undefined {
        return this.#blockNumber;
    }

    getTransactionId(): string | undefined {
        return this.#transactionId;
    }
}
