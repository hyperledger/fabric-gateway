/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * Allows access to information about a transaction that is committed to the ledger.
 */
public interface Commit extends Signable {
    /**
     * Get the transaction ID.
     * @return A transaction ID.
     */
    String getTransactionId();

    /**
     * Get the status of the committed transaction. If the transaction has not yet committed, this method blocks until
     * the commit occurs.
     * @return Transaction commit status.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    Status getStatus();
}
