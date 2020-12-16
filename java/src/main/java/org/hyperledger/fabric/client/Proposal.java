/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Map;

public interface Proposal {
    /**
     * Get the transaction ID.
     * @return A transaction ID.
     */
    String getTransactionId();

    /**
     * Get the serialized proposal message bytes.
     * @return A serialized proposal.
     */
    byte[] getBytes();

    /**
     * Get the digest of the serialized proposal. This is used to generate a digital signature.
     * @return A hash of the proposal.
     */
    byte[] getDigest();

    /**
     * Add transactions arguments to the proposal. These extend any previously added arguments.
     * @param args Transaction arguments.
     * @return This proposal.
     */
    Proposal addArguments(byte[]... args);

    /**
     * Add transactions arguments to the proposal. These extend any previously added arguments.
     * @param args Transaction arguments.
     * @return This proposal.
     */
    Proposal addArguments(String... args);

    /**
     * Associates all of the specified transient data keys and values with the proposal.
     * @param transientData Transient data keys and values.
     * @return This proposal.
     */
    Proposal putAllTransient(Map<String, byte[]> transientData);

    /**
     * Associates the specified transient data key and value with the proposal.
     * @param key Key with which the specified value is to be associated.
     * @param value Value to be associated with the specified key.
     * @return This proposal.
     */
    Proposal putTransient(String key, byte[] value);

    /**
     * Evaluate the proposal and return the transaction result. The transaction is not submitted to the orderer and is
     * not committed to the ledger.
     * @return Transaction result.
     */
    byte[] evaluate();

    /**
     * Send the proposal to peers to obtain endorsements. Successful endorsement results in a transaction that can be
     * submitted to the orderer to be committer to the ledger.
     * @return An endorsed transaction.
     */
    Transaction endorse();
}
