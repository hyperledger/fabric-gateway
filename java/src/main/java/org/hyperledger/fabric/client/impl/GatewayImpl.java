/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.security.GeneralSecurityException;
import java.util.concurrent.TimeUnit;

import io.grpc.Channel;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.GatewayRuntimeException;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.gateway.GatewayGrpc;

public final class GatewayImpl implements Gateway {
    public static final class Builder implements Gateway.Builder {
        private static final Signer UNDEFINED_SIGNER = (byte[] digest) -> {
            throw new UnsupportedOperationException("No signing implementation supplied");
        };
        private static final Runnable NO_OP_CLOSER = () -> {};

        private GatewayGrpc.GatewayBlockingStub service;
        private Runnable channelCloser = NO_OP_CLOSER;
        private Identity identity;
        private Signer signer = UNDEFINED_SIGNER; // No signer implementation is required if only offline signing is used

        @Override
        public Builder endpoint(final String target) { // TODO: Maybe should be abstracted out to Endpoint class
            ManagedChannel channel = ManagedChannelBuilder.forTarget(target).usePlaintext().build();
            service = GatewayGrpc.newBlockingStub(channel);
            channelCloser = () -> GatewayUtils.shutdownChannel(channel, 5, TimeUnit.SECONDS);
            return this;
        }

        @Override
        public Gateway.Builder connection(final Channel grpcChannel) {
            this.service = GatewayGrpc.newBlockingStub(grpcChannel);
            this.channelCloser = NO_OP_CLOSER;
            return this;
        }

        @Override
        public Builder identity(final Identity identity) {
            this.identity = identity;
            return this;
        }

        @Override
        public Builder signer(final Signer signer) {
            this.signer = signer;
            return this;
        }

        @Override
        public GatewayImpl connect() {
            return new GatewayImpl(this);
        }
    }

    private final GatewayGrpc.GatewayBlockingStub service;
    private final Runnable channelCloser;
    private final Identity identity;
    private final Signer signer;

    private GatewayImpl(final Builder builder) {
        this.identity = builder.identity;
        this.signer = builder.signer;
        this.service = builder.service;
        this.channelCloser = builder.channelCloser;

        if (null == this.identity) {
            throw new IllegalStateException("No client identity supplied");
        }
        if (null == this.service) {
            throw new IllegalStateException("No connections details supplied");
        }
    }

    @Override
    public void close() {
        channelCloser.run();
    }

    @Override
    public Network getNetwork(final String networkName) {
        return new NetworkImpl(networkName, this);
    }

    public Identity getIdentity() {
        return identity;
    }

    public byte[] sign(byte[] digest) {
        try {
            return signer.sign(digest);
        } catch (GeneralSecurityException e) {
            throw new GatewayRuntimeException(e);
        }
    }

    public byte[] hash(byte[] message) {
        return Hash.sha256(message);
    }

    public GatewayGrpc.GatewayBlockingStub getService() {
        return service;
    }
}
