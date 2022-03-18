/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEvent } from './chaincodeevent';

/**
 * Used to get the checkpointed state.
 */
export interface Checkpoint {
    /**
     * Get the checkpointed block number, or undefined if there is no previously saved state.
     */
    getBlockNumber(): bigint | undefined;

    /**
     * Get the last processed transaction ID within the current block.
     */
    getTransactionId(): string | undefined;
}

export interface Checkpointer extends Checkpoint {
    /**
     * To checkpoint block.
     */
    checkpointBlock(blockNumber: bigint): Promise<void>;

    /**
     *To checkpoint transaction within the current block.
     */
    checkpointTransaction(blockNumber: bigint, transactionId: string): Promise<void>;

    /**
     * To checkpoint chaincode event.
     */
    checkpointChaincodeEvent(event: ChaincodeEvent): Promise<void>;
}
