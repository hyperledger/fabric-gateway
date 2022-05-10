/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Optional;
import java.util.OptionalLong;

/**
 * Checkpoint provides the current position for event processing.
 */
public interface Checkpoint {
    /**
     * The block number in which the next event is expected.
     * @return  A ledger block number.
     */
    OptionalLong getBlockNumber();

    /**
     * Transaction Id of the last successfully processed event within the current block.
     * @return A transaction Id.
     */
    Optional<String> getTransactionId();
}
