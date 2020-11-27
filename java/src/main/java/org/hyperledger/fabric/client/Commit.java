/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.concurrent.Callable;

/**
 * A function that blocks until a transaction has been committed.
 */
@FunctionalInterface
public interface Commit extends Callable<byte[]> {
    /**
     * Wait for the transaction to be committed.
     * @return Transaction result.
     * @throws ContractException if the commit fails or timeout waiting for the commit is exceeded.
     */
    @Override
    byte[] call() throws ContractException; // TODO: Differentiate failure condition by exception type
}
