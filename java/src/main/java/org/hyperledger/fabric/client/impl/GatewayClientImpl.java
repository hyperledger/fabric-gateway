/*
 *  Copyright 2020 IBM All Rights Reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.util.Iterator;
import java.util.concurrent.TimeUnit;

import io.grpc.Channel;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.hyperledger.fabric.gateway.Event;
import org.hyperledger.fabric.gateway.GatewayGrpc;
import org.hyperledger.fabric.gateway.PreparedTransaction;
import org.hyperledger.fabric.gateway.ProposedTransaction;
import org.hyperledger.fabric.gateway.Result;

final class GatewayClientImpl implements GatewayClient {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final Runnable closer;

    public static GatewayClientImpl fromChannel(Channel channel) {
        GatewayGrpc.GatewayBlockingStub client = GatewayGrpc.newBlockingStub(channel);
        return new GatewayClientImpl(client, () -> {});
    }

    public static GatewayClientImpl fromEndpoint(String target) {
        ManagedChannel channel = ManagedChannelBuilder.forTarget(target).usePlaintext().build();
        GatewayGrpc.GatewayBlockingStub client = GatewayGrpc.newBlockingStub(channel);
        Runnable channelCloser = () -> {
            try {
                channel.shutdownNow().awaitTermination(5, TimeUnit.SECONDS);
            } catch (InterruptedException e) {
                // Ignore
            }
        };
        return new GatewayClientImpl(client, channelCloser);
    }

    private GatewayClientImpl(GatewayGrpc.GatewayBlockingStub client, Runnable closer) {
        this.client = client;
        this.closer = closer;
    }

    @Override
    public PreparedTransaction endorse(ProposedTransaction request) {
        return client.endorse(request);
    }

    @Override
    public Iterator<Event> submit(PreparedTransaction request) {
        return client.submit(request);
    }

    @Override
    public Result evaluate(ProposedTransaction request) {
        return client.evaluate(request);
    }

    @Override
    public void close() {
        closer.run();
    }
}
