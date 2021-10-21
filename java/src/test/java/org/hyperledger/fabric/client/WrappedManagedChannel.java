/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;
import java.util.concurrent.TimeUnit;

import io.grpc.CallOptions;
import io.grpc.ClientCall;
import io.grpc.ManagedChannel;
import io.grpc.MethodDescriptor;

/**
 * Wrapper for an existing managed channel to allow mocking of final channel implementation classes.
 */
public class WrappedManagedChannel extends ManagedChannel {
    private final ManagedChannel channel;

    public WrappedManagedChannel(final ManagedChannel channel) {
        Objects.requireNonNull(channel);
        this.channel = channel;
    }

    @Override
    public ManagedChannel shutdown() {
        return channel.shutdown();
    }

    @Override
    public boolean isShutdown() {
        return channel.isShutdown();
    }

    @Override
    public boolean isTerminated() {
        return channel.isTerminated();
    }

    @Override
    public ManagedChannel shutdownNow() {
        return channel.shutdownNow();
    }

    @Override
    public boolean awaitTermination(final long l, final TimeUnit timeUnit) throws InterruptedException {
        return channel.awaitTermination(l, timeUnit);
    }

    @Override
    public <RequestT, ResponseT> ClientCall<RequestT, ResponseT> newCall(final MethodDescriptor<RequestT, ResponseT> methodDescriptor, final CallOptions callOptions) {
        return channel.newCall(methodDescriptor, callOptions);
    }

    @Override
    public String authority() {
        return channel.authority();
    }
}
