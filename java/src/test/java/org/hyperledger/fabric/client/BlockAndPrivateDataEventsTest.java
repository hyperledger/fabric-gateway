/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.stream.Stream;

import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.ledger.rwset.Rwset;
import org.hyperledger.fabric.protos.peer.EventsPackage;

import static org.mockito.Mockito.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;

public final class BlockAndPrivateDataEventsTest extends CommonBlockEventsTest<EventsPackage.BlockAndPrivateData> {
    @Override
    protected void setEventsOptions(final Gateway.Builder builder, final CallOption... options) {
        builder.blockAndPrivateDataEventsOptions(options);
    }

    @Override
    protected EventsPackage.DeliverResponse newDeliverResponse(final long blockNumber) {
        return EventsPackage.DeliverResponse.newBuilder()
                .setBlockAndPrivateData(EventsPackage.BlockAndPrivateData.newBuilder()
                        .setBlock(Common.Block.newBuilder()
                                .setHeader(Common.BlockHeader.newBuilder().setNumber(blockNumber))
                        )
                        .putPrivateDataMap(0, Rwset.TxPvtReadWriteSet.newBuilder().build())
                )
                .build();
    }

    @Override
    protected void stubDoThrow(final Throwable... t) {
        doThrow(t).when(stub).blockAndPrivateDataEvents(any());
    }

    @Override
    protected CloseableIterator<EventsPackage.BlockAndPrivateData> getEvents(final CallOption... options) {
        return network.getBlockAndPrivateDataEvents(options);
    }

    @Override
    protected Stream<Common.Envelope> captureEvents() {
        return mocker.captureBlockAndPrivateDataEvents();
    }

    @Override
    protected EventsBuilder<EventsPackage.BlockAndPrivateData> newEventsRequest() {
        return network.newBlockAndPrivateDataEventsRequest();
    }

    @Override
    protected void stubDoReturn(final Stream<EventsPackage.DeliverResponse> responses) {
        doReturn(responses).when(stub).blockAndPrivateDataEvents(any());
    }

    @Override
    protected EventsPackage.BlockAndPrivateData extractEvent(final EventsPackage.DeliverResponse response) {
        return response.getBlockAndPrivateData();
    }
}

