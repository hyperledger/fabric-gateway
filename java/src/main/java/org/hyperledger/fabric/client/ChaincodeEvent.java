/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Arrays;
import java.util.Objects;

import org.hyperledger.fabric.protos.peer.ChaincodeEventPackage;

/**
 * Chaincode event emitted by a transaction function.
 */
public final class ChaincodeEvent {
    private final long blockNumber;
    private final String transactionId;
    private final String chaincodeId;
    private final String eventName;
    private final byte[] payload;
    private final int hash;

    ChaincodeEvent(final long blockNumber, final ChaincodeEventPackage.ChaincodeEvent event) {
        this.blockNumber = blockNumber;
        this.transactionId = event.getTxId();
        this.chaincodeId = event.getChaincodeId();
        this.eventName = event.getEventName();
        this.payload = event.getPayload().toByteArray();
        this.hash = Objects.hash(blockNumber, transactionId, chaincodeId, eventName); // Ignore potentially large payload; this is good enough
    }

    /**
     * Block number that included this chaincode event.
     * <p>Note that the block number is an unsigned 64-bit integer, with the sign bit used to hold the top bit of
     * the number.</p>
     * @return A block number.
     */
    public long getBlockNumber() {
        return blockNumber;
    }

    /**
     * Transaction that emitted this chaincode event.
     * @return Transaction ID.
     */
    public String getTransactionId() {
        return transactionId;
    }

    /**
     * Chaincode that emitted this event.
     * @return Chaincode ID.
     */
    public String getChaincodeId() {
        return chaincodeId;
    }

    /**
     * Name of the emitted event.
     * @return Event name.
     */
    public String getEventName() {
        return eventName;
    }

    /**
     * Application defined payload data associated with this event.
     * @return Event payload.
     */
    public byte[] getPayload() {
        return payload;
    }

    @Override
    public boolean equals(final Object other) {
        if (!(other instanceof ChaincodeEvent)) {
            return false;
        }

        ChaincodeEvent that = (ChaincodeEvent) other;

        return this.blockNumber == that.blockNumber
                && Objects.equals(this.transactionId, that.transactionId)
                && Objects.equals(this.chaincodeId, that.chaincodeId)
                && Objects.equals(this.eventName, that.eventName)
                && Arrays.equals(this.payload, that.payload);
    }

    @Override
    public int hashCode() {
        return hash;
    }

    @Override
    public String toString() {
        return GatewayUtils.toString(this,
                "blockNumber: " + blockNumber,
                "transactionId: " + transactionId,
                "chaincodeId: " + chaincodeId,
                "eventName: " + eventName,
                "payload: " + Arrays.toString(payload));
    }
}
