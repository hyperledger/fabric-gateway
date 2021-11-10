/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.StatusRuntimeException;

/**
 * Thrown when a failure occurs invoking a transaction.
 */
public class TransactionException extends GatewayException {
    private static final long serialVersionUID = 1L;

    private final String transactionId;

    /**
     * Constructs a new exception with the specified cause.
     * @param transactionId a transaction ID.
     * @param cause the cause.
     */
    public TransactionException(final String transactionId, final StatusRuntimeException cause) {
        super(cause);
        this.transactionId = transactionId;
    }

    /**
     * The ID of the transaction.
     * @return a transaction ID.
     */
    public String getTransactionId() {
        return transactionId;
    }
}
