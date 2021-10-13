/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * Chaincode event emitted by a transaction function.
 */
public interface ChaincodeEvent {
    /**
     * Block number that included this chaincode event.
     * <p>Note that the block number is an unsigned 64-bit integer, with the sign bit used to hold the top bit of
     * the number.</p>
     * @return A block number.
     */
    long getBlockNumber();

    /**
     * Transaction that emitted this chaincode event.
     * @return Transaction ID.
     */
    String getTransactionId();

    /**
     * Chaincode that emitted this event.
     * @return Chaincode name.
     */
    String getChaincodeName();

    /**
     * Name of the emitted event.
     * @return Event name.
     */
    String getEventName();

    /**
     * Application defined payload data associated with this event.
     * @return Event payload.
     */
    byte[] getPayload();
}
