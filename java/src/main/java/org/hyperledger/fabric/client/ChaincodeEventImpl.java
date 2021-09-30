/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Arrays;
import java.util.Objects;

import org.hyperledger.fabric.protos.peer.ChaincodeEventPackage;

final class ChaincodeEventImpl implements ChaincodeEvent {
    private final long blockNumber;
    private final String transactionId;
    private final String chaincodeId;
    private final String eventName;
    private final byte[] payload;
    private final int hash;

    ChaincodeEventImpl(final long blockNumber, final ChaincodeEventPackage.ChaincodeEvent event) {
        this.blockNumber = blockNumber;
        this.transactionId = event.getTxId();
        this.chaincodeId = event.getChaincodeId();
        this.eventName = event.getEventName();
        this.payload = event.getPayload().toByteArray();
        this.hash = Objects.hash(blockNumber, transactionId, chaincodeId, eventName); // Ignore potentially large payload; this is good enough
    }

    @Override
    public long getBlockNumber() {
        return blockNumber;
    }

    @Override
    public String getTransactionId() {
        return transactionId;
    }

    @Override
    public String getChaincodeId() {
        return chaincodeId;
    }

    @Override
    public String getEventName() {
        return eventName;
    }

    @Override
    public byte[] getPayload() {
        return payload;
    }

    @Override
    public boolean equals(final Object other) {
        if (!(other instanceof ChaincodeEventImpl)) {
            return false;
        }

        ChaincodeEventImpl that = (ChaincodeEventImpl) other;

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
