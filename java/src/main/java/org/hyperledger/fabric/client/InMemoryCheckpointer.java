/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * InMemoryCheckpointer class used to persist checkpointer state in memory.
 */
public final class InMemoryCheckpointer implements Checkpointer {

    private long blockNumber;
    private String transactionId = "";

    @Override
    public void checkpointBlock(final long blockNumber) {
        this.blockNumber = blockNumber + 1;
        this.transactionId = "";
    }

    @Override
    public void checkpointTransaction(final long blockNumber, final String transactionID) {
        this.blockNumber = blockNumber;
        this.transactionId = transactionID;
    }

    @Override
    public void  checkpointChaincodeEvent(final ChaincodeEvent event) {
        checkpointTransaction(event.getBlockNumber(), event.getTransactionId());
    }

    @Override
    public long getBlockNumber() {
        return this.blockNumber;
    }

    @Override
    public String getTransactionId() {
        return this.transactionId;
    }
}
