/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.peer.TxValidationCode;

/**
 * Thrown when a transaction fails to commit successfully.
 */
public class CommitException extends Exception {
    private static final long serialVersionUID = 1L;

    private final transient Status status;

    private static String message(final Status status) {
        TxValidationCode code = status.getCode();
        return "Commit of transaction " + status.getTransactionId() + " failed with status code " + code.getNumber()
                + " (" + code.name() + ")";
    }

    /**
     * Constructs a new commit exception for the specified transaction.
     * @param status Transaction commit status.
     */
    CommitException(final Status status) {
        super(message(status));
        this.status = status;
    }

    /**
     * Get the ID of the transaction.
     * @return transaction ID.
     */
    public String getTransactionId() {
        return status.getTransactionId();
    }

    /**
     * Get the transaction status code.
     * @return transaction status code.
     */
    public TxValidationCode getCode() {
        return status.getCode();
    }
}
