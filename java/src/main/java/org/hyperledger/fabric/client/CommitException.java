/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.peer.TransactionPackage;

/**
 * Thrown when a transaction fails to commit successfully.
 */
public final class CommitException extends Exception {
    private static final long serialVersionUID = 1L;

    private final String transactionId;
    private final TransactionPackage.TxValidationCode status;

    private static String message(final String transactionId, final TransactionPackage.TxValidationCode status) {
        return "Commit of transaction " + transactionId + " failed with status code " + status.getNumber()
                + " (" + status.name() + ")";
    }

    /**
     * Constructs a new commit exception for the specified transaction.
     * @param transactionId The ID of the transaction.
     * @param status Transaction validation code.
     */
    public CommitException(final String transactionId, final TransactionPackage.TxValidationCode status) {
        super(message(transactionId, status));
        this.transactionId = transactionId;
        this.status = status;
    }

    /**
     * Constructs a new commit exception for the specified transaction.
     * @param transactionId The ID of the transaction.
     * @param status Transaction validation code.
     * @param cause the cause.
     */
    public CommitException(final String transactionId, final TransactionPackage.TxValidationCode status, final Throwable cause) {
        super(message(transactionId, status), cause);
        this.transactionId = transactionId;
        this.status = status;
    }

    /**
     * Get the ID of the transaction.
     * @return transaction ID.
     */
    public String getTransactionId() {
        return transactionId;
    }

    /**
     * Get the commit status code for the transaction.
     * @return validation code.
     */
    public TransactionPackage.TxValidationCode getStatus() {
        return status;
    }
}
