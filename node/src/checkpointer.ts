/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CloseableAsyncIterable } from './client';

export interface Checkpointer {

    /**
     * Checkpoint the block number and transaction ID of an event. Checkpointing a different block number from the one
     * currently stored clears all previous transaction IDs.
     * @param blockNumber - a block number.
     * @param transactionId - a transaction ID.
     */
    checkpoint(blockNumber: bigint, transactionId?: string): Promise<void>;

    /**
     * Get the current block number, or undefined if there is no previously saved state.
     */
    getBlockNumber(): bigint | undefined;

    /**
     * Get the last processed transaction Id within the current block.
     */
    getTransactionId():string | undefined;
}

/**
 * An async iterable that can checkpoint events after they are processed.
 */
export interface CheckpointAsyncIterable<T> extends CloseableAsyncIterable<T> {
    /**
     * Checkpoint the last read event. This should be called immediately after the event is successfully processed.
     */
    checkpoint(): Promise<void>;
}
