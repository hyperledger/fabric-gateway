/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.io.InputStream;
import java.util.concurrent.TimeUnit;
import java.util.function.Function;

import io.grpc.Channel;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.netty.shaded.io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;

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
        private Function<byte[], byte[]> hash = Hash::sha256;
        private String target;
        private InputStream tlsRootCerts;
        private String serverNameOverride;
        private boolean useTLS = false;

        @Override
        // checkstyle:ignore-next-line:TodoComment
        public Builder endpoint(final String target) { // TODO: Maybe should be abstracted out to Endpoint class
            this.target = target;
            return this;
        }

        @Override
        public Builder tls(final InputStream tlsRootCerts, final String serverNameOverride) {
            this.useTLS = true;
            this.tlsRootCerts = tlsRootCerts;
            this.serverNameOverride = serverNameOverride;
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
        public Builder hash(final Function<byte[], byte[]> hash) {
            this.hash = hash;
            return this;
        }

        @Override
        public GatewayImpl connect() throws Exception {
            try {
                if (this.client == null) {
                    if (target == null) {
                        throw new IllegalStateException("No endpoint or channel has been provided");
                    }
                    ManagedChannel channel;
                    if (this.useTLS) {
                        channel = NettyChannelBuilder.forTarget(this.target)
                                .sslContext(GrpcSslContexts.forClient().trustManager(tlsRootCerts).build())
                                .overrideAuthority(serverNameOverride)
                                //.usePlaintext()
                                .build();
                    } else {
                        channel = ManagedChannelBuilder.forTarget(target).usePlaintext().build();
                    }
                    client = GatewayGrpc.newBlockingStub(channel);
                    channelCloser = () -> GatewayUtils.shutdownChannel(channel, CHANNEL_CLOSE_TIMEOUT.getTime(), CHANNEL_CLOSE_TIMEOUT.getTimeUnit());
                }
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

        this.signingIdentity = new SigningIdentity(builder.identity, builder.hash, builder.signer);
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

    // for test purposes
    Channel getChannel() {
        return client.getChannel();
    }
}
