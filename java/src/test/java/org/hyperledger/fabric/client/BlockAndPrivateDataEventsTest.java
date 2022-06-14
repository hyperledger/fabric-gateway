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
import org.hyperledger.fabric.protos.ledger.rwset.TxPvtReadWriteSet;
import org.hyperledger.fabric.protos.peer.BlockAndPrivateData;
import org.hyperledger.fabric.protos.peer.DeliverResponse;

import static org.mockito.Mockito.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;

public final class BlockAndPrivateDataEventsTest extends CommonBlockEventsTest<BlockAndPrivateData> {
    @Override
    protected void setEventsOptions(final Gateway.Builder builder, final UnaryOperator<CallOptions> options) {
        builder.blockAndPrivateDataEventsOptions(options);
    }

    @Override
    protected DeliverResponse newDeliverResponse(final long blockNumber) {
        return DeliverResponse.newBuilder()
                .setBlockAndPrivateData(BlockAndPrivateData.newBuilder()
                        .setBlock(Block.newBuilder()
                                .setHeader(BlockHeader.newBuilder().setNumber(blockNumber))
                        )
                        .putPrivateDataMap(0, TxPvtReadWriteSet.newBuilder().build())
                )
                .build();
    }

    @Override
    protected void stubDoThrow(final Throwable... t) {
        doThrow(t).when(stub).blockAndPrivateDataEvents(any());
    }

    @Override
    protected CloseableIterator<BlockAndPrivateData> getEvents() {
        return network.getBlockAndPrivateDataEvents();
    }

    @Override
    protected CloseableIterator<BlockAndPrivateData> getEvents(final UnaryOperator<CallOptions> options) {
        return network.getBlockAndPrivateDataEvents(options);
    }

    @Override
    protected Stream<Envelope> captureEvents() {
        return mocker.captureBlockAndPrivateDataEvents();
    }

    @Override
    protected EventsBuilder<BlockAndPrivateData> newEventsRequest() {
        return network.newBlockAndPrivateDataEventsRequest();
    }

    @Override
    protected void stubDoReturn(final Stream<DeliverResponse> responses) {
        doReturn(responses).when(stub).blockAndPrivateDataEvents(any());
    }

    @Override
    protected BlockAndPrivateData extractEvent(final DeliverResponse response) {
        return response.getBlockAndPrivateData();
    }
}

