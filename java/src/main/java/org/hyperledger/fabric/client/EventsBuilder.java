/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * Builder used to create a new events request.
 *
 * <p>If both a start block and checkpoint are specified, and the checkpoint has a valid position set, the
 * checkpoint position is used and the specified start block is ignored. If the checkpoint is unset then the start block
 * is used.</p>
 *
 * <p>If no start position is specified, eventing begins from the next committed block.</p>
 *
 * @param <T> Event type returned by the request.
 */
public interface EventsBuilder<T> extends Builder<EventsRequest<T>> {
    /**
     * Specify the block number at which to start reading events.
     * <p>Note that the block number is an unsigned 64-bit integer, with the sign bit used to hold the top bit of
     * the number.</p>
     * @param blockNumber a ledger block number.
     * @return This builder.
     */
    EventsBuilder<T> startBlock(long blockNumber);

    /**
     * Reads events starting at the checkpoint position.
     * @param checkpoint a checkpoint position.
     * @return This builder.
     */
    EventsBuilder<T> checkpoint(Checkpoint checkpoint);
}
