/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.peer.TxValidationCode;

/**
 * Information about a transaction that is committed to the ledger.
 */
public interface Status {
    /**
     * Get the transaction ID.
     * @return A transaction ID.
     */
    String getTransactionId();

    /**
     * Get the block number in which the transaction committed.
     * <p>Note that the block number is an unsigned 64-bit integer, with the sign bit used to hold the top bit of
     * the number.</p>
     * @return A block number.
     */
    long getBlockNumber();

    /**
     * Get the committed transaction status code.
     * @return Transaction status code.
     */
    TxValidationCode getCode();

    /**
     * Check whether the transaction committed successfully.
     * @return {@code true} if the transaction committed successfully; otherwise {@code false}.
     */
    boolean isSuccessful();
}
