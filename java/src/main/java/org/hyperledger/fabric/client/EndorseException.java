/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.StatusRuntimeException;

/**
 * Thrown when a failure occurs endorsing a transaction proposal.
 */
public class EndorseException extends TransactionException {
    private static final long serialVersionUID = 1L;

    /**
     * Constructs a new exception with the specified cause.
     * @param transactionId a transaction ID.
     * @param cause the cause.
     */
    public EndorseException(final String transactionId, final StatusRuntimeException cause) {
        super(transactionId, cause);
    }
}
