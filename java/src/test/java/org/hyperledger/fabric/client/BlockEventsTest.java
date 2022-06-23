/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.UnaryOperator;
import java.util.stream.Stream;

import io.grpc.CallOptions;
import org.hyperledger.fabric.protos.common.Block;
import org.hyperledger.fabric.protos.common.BlockHeader;
import org.hyperledger.fabric.protos.common.Envelope;
import org.hyperledger.fabric.protos.peer.DeliverResponse;

import static org.mockito.Mockito.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;

public final class BlockEventsTest extends CommonBlockEventsTest<Block> {
    @Override
    protected void setEventsOptions(final Gateway.Builder builder, final UnaryOperator<CallOptions> options) {
        builder.blockEventsOptions(options);
    }

    @Override
    protected DeliverResponse newDeliverResponse(final long blockNumber) {
        return DeliverResponse.newBuilder()
                .setBlock(Block.newBuilder()
                        .setHeader(BlockHeader.newBuilder().setNumber(blockNumber))
                )
                .build();
    }

    @Override
    protected void stubDoThrow(final Throwable... t) {
        doThrow(t).when(stub).blockEvents(any());
    }

    @Override
    protected CloseableIterator<Block> getEvents() {
        return network.getBlockEvents();
    }

    @Override
    protected CloseableIterator<Block> getEvents(final UnaryOperator<CallOptions> options) {
        return network.getBlockEvents(options);
    }

    @Override
    protected Stream<Envelope> captureEvents() {
        return mocker.captureBlockEvents();
    }

    @Override
    protected EventsBuilder<Block> newEventsRequest() {
        return network.newBlockEventsRequest();
    }

    @Override
    protected void stubDoReturn(final Stream<DeliverResponse> responses) {
        doReturn(responses).when(stub).blockEvents(any());
    }

    @Override
    protected Block extractEvent(final DeliverResponse response) {
        return response.getBlock();
    }
}

