/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * A Fabric Gateway call to obtain chaincode events. Supports off-line signing flow using
 * {@link Gateway#newSignedChaincodeEventsRequest(byte[], byte[])}.
 */
public interface ChaincodeEventsRequest extends EventsRequest<ChaincodeEvent> {
    /**
     * Builder used to create a new chaincode events request. The default behavior is to read events from the next
     * committed block.
     */
    interface Builder extends EventsBuilder<ChaincodeEvent> {
        @Override
        Builder startBlock(long blockNumber);

        @Override
        Builder checkpoint(Checkpoint checkpoint);

        @Override
        ChaincodeEventsRequest build();
    }
}
