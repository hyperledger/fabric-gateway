/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.io.IOException;

/**
 * Checkpointer allows update of a checkpoint position after events are successfully processed.
 */
public interface Checkpointer extends Checkpoint {
     /**
      * Checkpoint a successfully processed block.
      * <p>Note that the block number is an unsigned 64-bit integer, with the sign bit used to hold the top bit of
      * the number.</p>
      * @param blockNumber a ledger block number.
      * @throws IOException if an I/O error occurs.
      */
    void checkpointBlock(long blockNumber) throws IOException;

     /**
      * Checkpoint a transaction within a block.
      * @param blockNumber a ledger block number.
      * @param transactionId transaction id within the block.
      * @throws IOException if an I/O error occurs.
      */
    void checkpointTransaction(long blockNumber, String transactionId) throws IOException;

     /**
      * Checkpoint a chaincode event.
      * @param event a chaincode event.
      * @throws IOException if an I/O error occurs.
      */
    void checkpointChaincodeEvent(ChaincodeEvent event) throws IOException;

}

