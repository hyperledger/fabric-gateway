/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.gateway.CommitStatusResponse;

import static org.hyperledger.fabric.protos.peer.TransactionPackage.TxValidationCode;

/**
 * Information about a transaction that is committed to the ledger.
 */
public final class Status {
    private final String transactionId;
    private final long blockNumber;
    private final TxValidationCode code;

    Status(final String transactionId, final CommitStatusResponse response) {
        this.transactionId = transactionId;
        this.blockNumber = response.getBlockNumber();
        this.code = response.getResult();
    }

    /**
     * Get the transaction ID.
     * @return A transaction ID.
     */
    public String getTransactionId() {
        return transactionId;
    }

    /**
     * Get the block number in which the transaction committed.
     * <p>Note that the block number is an unsigned 64-bit integer, with the sign bit used to hold the top bit of
     * the number.</p>
     * @return A block number.
     */
    public long getBlockNumber() {
        return blockNumber;
    }

    /**
     * Get the committed transaction status code.
     * @return Transaction status code.
     */
    public TxValidationCode getCode() {
        return code;
    };

    /**
     * Check whether the transaction committed successfully.
     * @return {@code true} if the transaction committed successfully; otherwise {@code false}.
     */
    public boolean isSuccessful() {
        return code == TxValidationCode.VALID;
    };
}
