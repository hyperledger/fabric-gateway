/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.common.Block;

/**
 * A Fabric Gateway call to obtain block events. Supports off-line signing flow using
 * {@link Gateway#newSignedBlockEventsRequest(byte[], byte[])}.
 */
public interface BlockEventsRequest extends EventsRequest<Block> {
    /**
     * Builder used to create a new block events request. The default behavior is to read events from the next
     * committed block.
     */
    interface Builder extends EventsBuilder<Block> {
        @Override
        Builder startBlock(long blockNumber);

        @Override
        Builder checkpoint(Checkpoint checkpoint);

        @Override
        BlockEventsRequest build();
    }
}
