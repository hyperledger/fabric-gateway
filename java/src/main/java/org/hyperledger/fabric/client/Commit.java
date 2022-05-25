/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.UnaryOperator;

import io.grpc.CallOptions;

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
     * @throws CommitStatusException if the gRPC service invocation fails.
     */
    default Status getStatus() throws CommitStatusException {
        return getStatus(GatewayUtils.asCallOptions());
    }

    /**
     * Get the status of the committed transaction. If the transaction has not yet committed, this method blocks until
     * the commit occurs.
     * @param options Function that transforms call options.
     * @return Transaction commit status.
     * @throws CommitStatusException if the gRPC service invocation fails.
     */
    Status getStatus(UnaryOperator<CallOptions> options) throws CommitStatusException;

    /**
     * Get the status of the committed transaction. If the transaction has not yet committed, this method blocks until
     * the commit occurs.
     * @param options Call options.
     * @return Transaction commit status.
     * @throws CommitStatusException if the gRPC service invocation fails.
     * @deprecated Replaced by {@link #getStatus(UnaryOperator)}.
     */
    @Deprecated
    default Status getStatus(CallOption... options) throws CommitStatusException {
        return getStatus(GatewayUtils.asCallOptions(options));
    }
}
