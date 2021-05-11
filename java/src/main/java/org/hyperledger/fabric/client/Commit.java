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
public interface Commit {
    /**
     * Get the serialized transaction message bytes.
     * @return A serialized transaction.
     */
    byte[] getBytes();

    /**
     * Get the digest of the serialized transaction. This is used to generate a digital signature.
     * @return A hash of the transaction.
     */
    byte[] getDigest();

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

}
