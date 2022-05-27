/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.UnaryOperator;
import java.util.stream.Stream;

import io.grpc.CallOptions;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.peer.EventsPackage;

import static org.mockito.Mockito.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;

public final class FilteredBlockEventsTest extends CommonBlockEventsTest<EventsPackage.FilteredBlock> {
    @Override
    protected void setEventsOptions(final Gateway.Builder builder, final UnaryOperator<CallOptions> options) {
        builder.filteredBlockEventsOptions(options);
    }

    @Override
    protected EventsPackage.DeliverResponse newDeliverResponse(final long blockNumber) {
        return EventsPackage.DeliverResponse.newBuilder()
                .setFilteredBlock(EventsPackage.FilteredBlock.newBuilder()
                        .setNumber(blockNumber)
                )
                .build();
    }

    @Override
    protected void stubDoThrow(final Throwable... t) {
        doThrow(t).when(stub).filteredBlockEvents(any());
    }

    @Override
    protected CloseableIterator<EventsPackage.FilteredBlock> getEvents() {
        return network.getFilteredBlockEvents();
    }

    @Override
    protected CloseableIterator<EventsPackage.FilteredBlock> getEvents(final UnaryOperator<CallOptions> options) {
        return network.getFilteredBlockEvents(options);
    }

    @Override
    protected Stream<Common.Envelope> captureEvents() {
        return mocker.captureFilteredBlockEvents();
    }

    @Override
    protected EventsBuilder<EventsPackage.FilteredBlock> newEventsRequest() {
        return network.newFilteredBlockEventsRequest();
    }

    @Override
    protected void stubDoReturn(final Stream<EventsPackage.DeliverResponse> responses) {
        doReturn(responses).when(stub).filteredBlockEvents(any());
    }

    @Override
    protected EventsPackage.FilteredBlock extractEvent(final EventsPackage.DeliverResponse response) {
        return response.getFilteredBlock();
    }
}

