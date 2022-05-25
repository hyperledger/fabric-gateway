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

public final class BlockEventsTest extends CommonBlockEventsTest<Common.Block> {
    @Override
    protected void setEventsOptions(final Gateway.Builder builder, final UnaryOperator<CallOptions> options) {
        builder.blockEventsOptions(options);
    }

    @Override
    protected EventsPackage.DeliverResponse newDeliverResponse(final long blockNumber) {
        return EventsPackage.DeliverResponse.newBuilder()
                .setBlock(Common.Block.newBuilder()
                        .setHeader(Common.BlockHeader.newBuilder().setNumber(blockNumber))
                )
                .build();
    }

    @Override
    protected void stubDoThrow(final Throwable... t) {
        doThrow(t).when(stub).blockEvents(any());
    }

    @Override
    protected CloseableIterator<Common.Block> getEvents() {
        return network.getBlockEvents();
    }

    @Override
    protected CloseableIterator<Common.Block> getEvents(final UnaryOperator<CallOptions> options) {
        return network.getBlockEvents(options);
    }

    @Override
    protected Stream<Common.Envelope> captureEvents() {
        return mocker.captureBlockEvents();
    }

    @Override
    protected EventsBuilder<Common.Block> newEventsRequest() {
        return network.newBlockEventsRequest();
    }

    @Override
    protected void stubDoReturn(final Stream<EventsPackage.DeliverResponse> responses) {
        doReturn(responses).when(stub).blockEvents(any());
    }

    @Override
    protected Common.Block extractEvent(final EventsPackage.DeliverResponse response) {
        return response.getBlock();
    }
}

