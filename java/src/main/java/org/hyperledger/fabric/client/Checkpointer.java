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
      */
      void  checkpointBlock(long blockNumber) throws IOException;

     /**
      * CheckpointTransaction checkpoints the transaction within a block.
      * @param blockNumber a ledger block number.
      * @param transactionId transaction id within the block.
      */
      void checkpointTransaction(long blockNumber, String transactionId) throws Exception;

     /**
      *  CheckpointChaincodeEvent checkpoints the chaincode event.
      * @param event a chaincode event.
      */
      void checkpointChaincodeEvent(ChaincodeEvent event) throws Exception;

}

