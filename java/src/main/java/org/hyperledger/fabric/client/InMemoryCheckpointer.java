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

    /**
     * CheckpointBlock checkpoints the block.
     * @param blockNumber
     */
    public void checkpointBlock(final long blockNumber) {
        this.blockNumber = blockNumber + 1;
        this.transactionId = "";
    }

    /**
     * CheckpointTransaction checkpoints the transaction within a block.
     * @param blockNumber
     * @param transactionID
     */
    public void checkpointTransaction(final long blockNumber, final String transactionID) {
        this.blockNumber = blockNumber;
        this.transactionId = transactionID;
    }

    /**
     * CheckpointChaincodeEvent checkpoints the chaincode event.
     * @param event
     */
    public void  checkpointChaincodeEvent(final ChaincodeEvent event) {
        checkpointTransaction(event.getBlockNumber(), event.getTransactionId());
    }

    /**
     * @return BlockNumber in which the next event is expected.
     */
    public long getBlockNumber() {
        return this.blockNumber;
    }

    /**
     * @return Transaction ID of the last successfully processed event within the current block.
     */
    public String getTransactionId() {
        return this.transactionId;
    }
}
