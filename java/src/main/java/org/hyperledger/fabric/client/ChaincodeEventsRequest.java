/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * A Fabric Gateway call to obtain chaincode events. Supports off-line signing flow using
 * {@link Network#newSignedChaincodeEventsRequest(byte[], byte[])}.
 */
public interface ChaincodeEventsRequest extends Signable {
    /**
     * Get events emitted by transaction functions of a specific chaincode. The Java gRPC implementation may not begin
     * reading events until the first use of the returned iterator.
     * <p>Note that the returned iterator may
     * throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC connection error occurs.</p>
     * @return Ordered sequence of events.
     */
    CloseableIterator<ChaincodeEvent> getEvents();

    /**
     * Builder used to create a new chaincode events request. The default behavior is to read chaincode events from the
     * next committed block.
     */
    interface Builder {
        /**
         * Specify the block number at which to start reading chaincode events.
         * <p>Note that the block number is an unsigned 64-bit integer, with the sign bit used to hold the top bit of
         * the number.</p>
         * @param blockNumber a ledger block number.
         * @return This builder.
         */
        Builder startBlock(long blockNumber);

        /**
         * Build the chaincode events request from the configuration state of this builder.
         * @return A chaincode events request.
         */
        ChaincodeEventsRequest build();
    }
}
