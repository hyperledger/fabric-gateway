/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.concurrent.TimeUnit;

import io.grpc.Channel;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.gateway.GatewayGrpc;

final class GatewayImpl implements Gateway {
    public static final class Builder implements Gateway.Builder {
        private static final Signer UNDEFINED_SIGNER = (digest) -> {
            throw new UnsupportedOperationException("No signing implementation supplied");
        };
        private static final Runnable NO_OP_CLOSER = () -> { };
        private static final TimePeriod CHANNEL_CLOSE_TIMEOUT = new TimePeriod(5, TimeUnit.SECONDS);

        private GatewayGrpc.GatewayBlockingStub client;
        private Runnable channelCloser = NO_OP_CLOSER;
        private Identity identity;
        private Signer signer = UNDEFINED_SIGNER; // No signer implementation is required if only offline signing is used

        @Override
        // checkstyle:ignore-next-line:TodoComment
        public Builder endpoint(final String target) { // TODO: Maybe should be abstracted out to Endpoint class
            ManagedChannel channel = ManagedChannelBuilder.forTarget(target).usePlaintext().build();
            client = GatewayGrpc.newBlockingStub(channel);
            channelCloser = () -> GatewayUtils.shutdownChannel(channel, CHANNEL_CLOSE_TIMEOUT.getTime(), CHANNEL_CLOSE_TIMEOUT.getTimeUnit());
            return this;
        }

        @Override
        public Gateway.Builder connection(final Channel grpcChannel) {
            this.client = GatewayGrpc.newBlockingStub(grpcChannel);
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
            try {
                return new GatewayImpl(this);
            } catch (Exception e) {
                channelCloser.run(); // Ensure no orphaned gRPC channels
                throw e;
            }
        }
    }

    private final GatewayGrpc.GatewayBlockingStub client;
    private final Runnable channelCloser;
    private final SigningIdentity signingIdentity;

    private GatewayImpl(final Builder builder) {
        if (null == builder.identity) {
            throw new IllegalStateException("No client identity supplied");
        }
        if (null == builder.client) {
            throw new IllegalStateException("No connections details supplied");
        }

        this.signingIdentity = new SigningIdentity(builder.identity, builder.signer);
        this.client = builder.client;
        this.channelCloser = builder.channelCloser;
    }

    @Override
    public Identity getIdentity() {
        return this.signingIdentity.getIdentity();
    }

    @Override
    public void close() {
        channelCloser.run();
    }

    @Override
    public Network getNetwork(final String networkName) {
        return new NetworkImpl(client, signingIdentity, networkName);
    }
}
