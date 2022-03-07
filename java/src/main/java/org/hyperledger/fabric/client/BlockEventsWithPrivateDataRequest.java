/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.peer.EventsPackage;

///**
// * A Fabric Gateway call to obtain block events with private data. Supports off-line signing flow using
// * {@link Gateway#newSignedBlockEventsWithPrivateDataRequest(byte[], byte[])}.
// */

/**
 * A Fabric Gateway call to obtain block events with private data.
 */
public interface BlockEventsWithPrivateDataRequest extends EventsRequest<EventsPackage.BlockAndPrivateData> {
    /**
     * Builder used to create a new block events with private data request. The default behavior is to read events from
     * the next committed block.
     */
    interface Builder extends EventsBuilder<EventsPackage.BlockAndPrivateData> {
        @Override
        Builder startBlock(long blockNumber);

        @Override
        BlockEventsWithPrivateDataRequest build();
    }
}
