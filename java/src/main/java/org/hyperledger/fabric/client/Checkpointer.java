/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * Checkpointer allows update of a checkpoint position after events are successfully processed.
 */
public interface Checkpointer extends Checkpoint {

     /**
      * CheckpointBlock checkpoints block.
      * @param blockNumber
      */
      void  checkpointBlock(long blockNumber) throws Exception;

     /**
      * CheckpointTransaction checkpoints the transaction within a block.
      * @param blockNumber
      * @param transactionId
      */
      void checkpointTransaction(long blockNumber, String transactionId) throws Exception;


     /**
      *  CheckpointChaincodeEvent checkpoints the chaincode event.
      * @param event
      */
      void checkpointChaincodeEvent(ChaincodeEvent event) throws Exception;


}

