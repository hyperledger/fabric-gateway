/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.StatusRuntimeException;

/**
 * Thrown when a failure occurs submitting an endorsed transaction to the orderer.
 */
public class SubmitException extends GatewayException {
    private static final long serialVersionUID = 1L;

    /**
     * Constructs a new exception with the specified cause.
     * @param cause the cause.
     */
    public SubmitException(final StatusRuntimeException cause) {
        super(cause);
    }
}
