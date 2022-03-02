/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CloseableAsyncIterable } from './client';

export interface Checkpointer {

    /**
    * To checkpoint the blocknumber and transaction ID of the event.
    */
    checkpoint(blockNumber: bigint, transactionId?: string): Promise<void>;

    /**
    * Get the current block number, or undefined if there is no previously saved state.
    */
    getBlockNumber(): bigint | undefined;
    /**
    * Get the transaction IDs processed within the current block.
    */
    getTransactionIds(): Set<string>;
}

/**
 * An async iterable that can be used to iterate over the events and checkpoint the events after processing.
 */
export interface CheckpointAsyncIterable<T> extends CloseableAsyncIterable<T> {
    /**
    * To checkpoint the events after processing.
    */
    checkpoint(): Promise<void>;
}
