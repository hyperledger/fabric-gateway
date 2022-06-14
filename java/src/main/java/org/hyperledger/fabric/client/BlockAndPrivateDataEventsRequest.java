/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.peer.BlockAndPrivateData;

/**
 * A Fabric Gateway call to obtain block and private data events. Supports off-line signing flow using
 * {@link Gateway#newSignedBlockAndPrivateDataEventsRequest(byte[], byte[])}.
 */
public interface BlockAndPrivateDataEventsRequest extends EventsRequest<BlockAndPrivateData> {
    /**
     * Builder used to create a new block and private data events request. The default behavior is to read events from
     * the next committed block.
     */
    interface Builder extends EventsBuilder<BlockAndPrivateData> {
        @Override
        Builder startBlock(long blockNumber);

        @Override
        Builder checkpoint(Checkpoint checkpoint);

        @Override
        BlockAndPrivateDataEventsRequest build();
    }
}
