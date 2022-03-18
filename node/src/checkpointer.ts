/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

export interface Checkpointer {
    /**
     * Get the block number for the next event, or undefined if there is no previously saved state.
     */
    getBlockNumber(): bigint | undefined;

    /**
     * Get the last processed transaction Id within the current block.
     */
    getTransactionId(): string | undefined;
}
