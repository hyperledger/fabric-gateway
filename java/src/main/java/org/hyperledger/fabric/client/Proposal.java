/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Map;

/**
 * A Fabric Gateway transaction proposal, which can be used to evaluate a transaction to query ledger state, or obtain
 * endorsements so that the transaction can be submitted to update ledger state. Supports off-line signing flow using
 * {@link Network#newSignedChaincodeEventsRequest(byte[], byte[])}.
 */
public interface Proposal extends Signable {
    /**
     * Get the transaction ID.
     * @return A transaction ID.
     */
    String getTransactionId();

    /**
     * Evaluate the proposal and return the transaction result. The transaction is not submitted to the orderer and is
     * not committed to the ledger.
     * @return Transaction result.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    byte[] evaluate();

    /**
     * Send the proposal to peers to obtain endorsements. Successful endorsement results in a transaction that can be
     * submitted to the orderer to be committer to the ledger.
     * @return An endorsed transaction.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    Transaction endorse();

    /**
     * Builder used to create a new transaction proposal.
     */
    interface Builder {
        /**
         * Add transactions arguments to the proposal. These extend any previously added arguments.
         * @param args Transaction arguments.
         * @return This builder.
         */
        Builder addArguments(byte[]... args);

        /**
         * Add transactions arguments to the proposal. These extend any previously added arguments.
         * @param args Transaction arguments.
         * @return This builder.
         */
        Builder addArguments(String... args);

        /**
         * Associates all of the specified transient data keys and values with the proposal.
         * @param transientData Transient data keys and values.
         * @return This builder.
         */
        Builder putAllTransient(Map<String, byte[]> transientData);

        /**
         * Associates the specified transient data key and value with the proposal.
         * @param key Key with which the specified value is to be associated.
         * @param value Value to be associated with the specified key.
         * @return This builder.
         */
        Builder putTransient(String key, byte[] value);

        /**
         * Associates the specified transient data key and value with the proposal.
         * @param key Key with which the specified value is to be associated.
         * @param value Value to be associated with the specified key.
         * @return This builder.
         */
        Builder putTransient(String key, String value);

        /**
         * Specifies the set of organizations that will attempt to endorse the proposal.
         * No other organizations' peers will be sent this proposal.
         * This is usually used in conjunction with {@link #putTransient(String, byte[])} or
         * {@link #putAllTransient(Map)} for private data scenarios.
         * @param mspids The Member Services Provider IDs of the endorsing organizations.
         * @return This builder.
         */
        Builder setEndorsingOrganizations(String... mspids);

        /**
         * Build the proposal from the configuration state of this builder. A new transaction ID will be generated on
         * each invocation of this method.
         * @return A proposal.
         */
        Proposal build();
    }
}
