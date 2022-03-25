/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.gson.Gson;

import java.io.FileNotFoundException;
import java.io.FileReader;
import java.io.FileWriter;
class CheckpointerState {
    private long blockNumber;
    private String transactionId;
    CheckpointerState(final long blockNumber, final String transactionId) {
        this.blockNumber = blockNumber;
        this.transactionId = transactionId;
    }

    public String getTransactionId() {
        return transactionId;
    }

    public long getBlockNumber() {
        return blockNumber;
    }
}

/**
 * FileCheckpointer to store checkpointer state during file read write operations.
 */
public final class FileCheckpointer implements Checkpointer {

    private long blockNumber;
    private String transactionId = "";
    private String path = "";

    FileCheckpointer(final String path) throws Exception {
        this.path = path;
        this.init();
    }

    private void init() throws Exception {
        this.loadFromFile();
        this.saveToFile();
    }

    /**
     * CheckpointBlock checkpoints the block.
     * @param blockNumber
     */
    public void checkpointBlock(final long blockNumber) throws Exception {
        this.blockNumber = blockNumber + 1;
        this.transactionId = "";
        this.saveToFile();
    }

    /**
     * CheckpointTransaction checkpoints the transaction within a block.
     * @param blockNumber
     * @param transactionID
     */
    public void checkpointTransaction(final long blockNumber, final String transactionID) throws Exception {
        this.blockNumber = blockNumber;
        this.transactionId = transactionID;
        this.saveToFile();
    }

    /**
     * CheckpointChaincodeEvent checkpoints the chaincode event.
     * @param event
     */
    public void  checkpointChaincodeEvent(final ChaincodeEvent event) throws Exception {
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

    private void loadFromFile() throws Exception {
            CheckpointerState state = this.readFile();
            this.setState(state);
    }

    private void saveToFile() throws Exception {
        CheckpointerState state = this.getState();

        Gson gson = new Gson();
        gson.toJson(state, new FileWriter(this.path));
    }

    private void setState(final CheckpointerState state) {
        this.blockNumber = state.getBlockNumber();
        this.transactionId = state.getTransactionId();
    }

    private CheckpointerState getState() {
        return new CheckpointerState(this.blockNumber, this.transactionId);
    }

    private CheckpointerState readFile() throws Exception {
        try {
            FileReader fileReader = new FileReader(this.path);
            Gson gson = new Gson();
            return  gson.fromJson(fileReader, CheckpointerState.class);
        } catch (FileNotFoundException e) {
            //ignore
        }
        return null;
    }
}
