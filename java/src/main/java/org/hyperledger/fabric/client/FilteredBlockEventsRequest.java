/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.peer.FilteredBlock;

/**
 * A Fabric Gateway call to obtain filtered block events. Supports off-line signing flow using
 * {@link Gateway#newSignedFilteredBlockEventsRequest(byte[], byte[])}.
 */
public interface FilteredBlockEventsRequest extends EventsRequest<FilteredBlock> {
    /**
     * Builder used to create a new filtered block events request. The default behavior is to read events from the next
     * committed block.
     */
    interface Builder extends EventsBuilder<FilteredBlock> {
        @Override
        Builder startBlock(long blockNumber);

        @Override
        Builder checkpoint(Checkpoint checkpoint);

        @Override
        FilteredBlockEventsRequest build();
    }
}
