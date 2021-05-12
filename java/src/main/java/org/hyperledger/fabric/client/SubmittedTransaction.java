/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * Allows access to the transaction result and its commit status on the ledger.
 */
public interface SubmittedTransaction extends Commit {
    /**
     * Get the transaction result. This is obtained during the endorsement process when the transaction proposal is
     * run on endorsing peers and so is available immediately. The transaction might subsequently fail to commit
     * successfully.
     * @return Transaction result.
     */
    byte[] getResult();
}
