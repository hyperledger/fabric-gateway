/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.UnaryOperator;

import io.grpc.CallOptions;

/**
 * An endorsed transaction that can be submitted to the orderer for commit to the ledger.
 */
public interface Transaction extends Signable {
    /**
     * Get the transaction result. The result is obtained as part of the proposal endorsement so may be read
     * immediately. It is not necessary to submit the transaction before getting the transaction result, but the
     * transaction will not be committed to the ledger and its effects visible to other clients and transactions until
     * after it has been submitted to the orderer.
     * @return A transaction result.
     */
    byte[] getResult();

    /**
     * Get the transaction ID.
     * @return A transaction ID.
     */
    String getTransactionId();

    /**
     * Submit the transaction to the orderer to be committed to the ledger. This method blocks until the transaction
     * has been successfully committed to the ledger.
     * @return A transaction result.
     * @throws SubmitException if the submit invocation fails.
     * @throws CommitStatusException if the commit status invocation fails.
     * @throws CommitException if the transaction commits unsuccessfully.
     */
    default byte[] submit() throws SubmitException, CommitStatusException, CommitException {
        return submit(GatewayUtils.asCallOptions());
    }

    /**
     * Submit the transaction to the orderer to be committed to the ledger. This method blocks until the transaction
     * has been successfully committed to the ledger.
     * @param options Function that transforms call options.
     * @return A transaction result.
     * @throws SubmitException if the submit invocation fails.
     * @throws CommitStatusException if the commit status invocation fails.
     * @throws CommitException if the transaction commits unsuccessfully.
     */
    byte[] submit(UnaryOperator<CallOptions> options) throws SubmitException, CommitStatusException, CommitException;

    /**
     * Submit the transaction to the orderer to be committed to the ledger. This method blocks until the transaction
     * has been successfully committed to the ledger.
     * @param options Call options.
     * @return A transaction result.
     * @throws SubmitException if the submit invocation fails.
     * @throws CommitStatusException if the commit status invocation fails.
     * @throws CommitException if the transaction commits unsuccessfully.
     * @deprecated Replaced by {@link #submit(UnaryOperator)}.
     */
    @Deprecated
    default byte[] submit(CallOption... options) throws SubmitException, CommitStatusException, CommitException {
        return submit(GatewayUtils.asCallOptions(options));
    }

    /**
     * Submit the transaction to the orderer to be committed to the ledger. This method returns immediately after the
     * transaction is successfully delivered to the orderer. The returned Commit may be used to subsequently wait
     * for the transaction to be committed to the ledger.
     * @return A transaction commit.
     * @throws SubmitException if the gRPC service invocation fails.
     */
    default SubmittedTransaction submitAsync() throws SubmitException {
        return submitAsync(GatewayUtils.asCallOptions());
    }

    /**
     * Submit the transaction to the orderer to be committed to the ledger. This method returns immediately after the
     * transaction is successfully delivered to the orderer. The returned Commit may be used to subsequently wait
     * for the transaction to be committed to the ledger.
     * @param options Function that transforms call options.
     * @return A transaction commit.
     * @throws SubmitException if the gRPC service invocation fails.
     */
    SubmittedTransaction submitAsync(UnaryOperator<CallOptions> options) throws SubmitException;

    /**
     * Submit the transaction to the orderer to be committed to the ledger. This method returns immediately after the
     * transaction is successfully delivered to the orderer. The returned Commit may be used to subsequently wait
     * for the transaction to be committed to the ledger.
     * @param options Call options.
     * @return A transaction commit.
     * @throws SubmitException if the gRPC service invocation fails.
     * @deprecated Replaced by {@link #submit(UnaryOperator)}.
     */
    @Deprecated
    default SubmittedTransaction submitAsync(CallOption... options) throws SubmitException {
        return submitAsync(GatewayUtils.asCallOptions(options));
    }
}
