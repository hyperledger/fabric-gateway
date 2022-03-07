/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.peer.EventsPackage;

///**
// * A Fabric Gateway call to obtain filtered block events. Supports off-line signing flow using
// * {@link Gateway#newSignedFilteredBlockEventsRequest(byte[], byte[])}.
// */

/**
 * A Fabric Gateway call to obtain filtered block events.
 */
public interface FilteredBlockEventsRequest extends EventsRequest<EventsPackage.FilteredBlock> {
    /**
     * Builder used to create a new filtered block events request. The default behavior is to read events from the next
     * committed block.
     */
    interface Builder extends EventsBuilder<EventsPackage.FilteredBlock> {
        @Override
        Builder startBlock(long blockNumber);

        @Override
        FilteredBlockEventsRequest build();
    }
}
