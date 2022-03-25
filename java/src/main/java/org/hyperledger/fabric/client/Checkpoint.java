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
     *
     * @return  BlockNumber in which the next event is expected.
     */
    long getBlockNumber();

    /**
     *
     * @return Transaction ID of the last successfully processed event within the current block.
     */
    String getTransactionId();
}
