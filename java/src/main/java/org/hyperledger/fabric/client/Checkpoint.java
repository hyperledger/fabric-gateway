/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * Checkpoint provides the current position for event processing.
 */
public interface Checkpoint {
    /**
     * The block number in which the next event is expected.
     * @return  A ledger block number.
     */
    long getBlockNumber();

    /**
     * Transaction Id of the last successfully processed event within the current block.
     * @return A transaction Id.
     */
    String getTransactionId();
}
