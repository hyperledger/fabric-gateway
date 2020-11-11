/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.gateway;

/**
 * Thrown when an error occurs invoking a smart contract.
 */
public class ContractException extends GatewayException {
    private static final long serialVersionUID = -1278679656087547825L;

    /**
     * Constructs a new exception with the specified detail message and cause. The payload is not initialized.
     * @param message the detail message.
     * @param cause the cause.
     */
    public ContractException(final String message, final Throwable cause) {
        super(message, cause);
    }

    /**
     * Constructs a new exception with the specified detail message and proposal responses returned from peer
     * invocations. The cause is not initialized.
     * @param message the detail message.
     */
    public ContractException(final String message) {
        super(message);
    }
}
