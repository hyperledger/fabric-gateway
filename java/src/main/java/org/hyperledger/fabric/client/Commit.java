/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.peer.TransactionPackage;

/**
 * Allows access to information about a transaction that is committed to the ledger.
 */
public interface Commit extends Signable {
    /**
     * Get the committed transaction status code. If the transaction has not yet committed, this method blocks until
     * the commit occurs.
     * @return Transaction commit status.
     */
    TransactionPackage.TxValidationCode getStatus();

    /**
     * Check whether the transaction committed successfully. If the transaction has not yet committed, this method
     * blocks until the commit occurs.
     * @return {@code true} if the transaction committed successfully; otherwise {@code false}.
     */
    boolean isSuccessful();

    /**
     * Get the transaction ID.
     * @return A transaction ID.
     */
    String getTransactionId();

    /**
     * Get the block number in which the transaction committed. If the transaction has not yet committed, this method
     * blocks until the commit occurs.
     * @return A block number.
     */
    long getBlockNumber();
}
