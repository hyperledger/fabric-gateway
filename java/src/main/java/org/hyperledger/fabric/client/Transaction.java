/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

public interface Transaction {
    /**
     * Get the transaction result. The result is obtained as part of the proposal endorsement so may be read
     * immediately. It is not necessary to submit the transaction before getting the transaction result, but the
     * transaction will not be committed to the ledger and its effects visible to other clients and transactions until
     * after it has been submitted to the orderer.
     * @return A transaction result.
     */
    byte[] getResult();

    /**
     * Get the serialized transaction message bytes.
     * @return A serialized transaction.
     */
    byte[] getBytes();

    /**
     * Get the digest of the serialized transaction. This is used to generate a digital signature.
     * @return A hash of the transaction.
     */
    byte[] getDigest();

    /**
     * Get the transaction ID.
     * @return A transaction ID.
     */
    String getTransactionId();

    /**
     * Submit the transaction to the orderer to be committed to the ledger. This method blocks until the transaction
     * has been successfully committed to the ledger.
     * @return A transaction result.
     * @throws CommitException if the transaction fails to commit successfully.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    byte[] submit() throws CommitException;

    /**
     * Submit the transaction to the orderer to be committed to the ledger. This method returns immediately after the
     * transaction is successfully delivered to the orderer. The returned Commit may be used to subsequently wait
     * for the transaction to be committed to the ledger.
     * @return A transaction commit.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    SubmittedTransaction submitAsync();
}
