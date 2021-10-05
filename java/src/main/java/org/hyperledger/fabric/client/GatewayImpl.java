/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.Function;

import io.grpc.Channel;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;

final class GatewayImpl implements Gateway {
    public static final class Builder implements Gateway.Builder {
        private static final Signer UNDEFINED_SIGNER = (digest) -> {
            throw new UnsupportedOperationException("No signing implementation supplied");
        };

        private GatewayClient client;
        private Identity identity;
        private Signer signer = UNDEFINED_SIGNER; // No signer implementation is required if only offline signing is used
        private Function<byte[], byte[]> hash = Hash::sha256;

        @Override
        public Gateway.Builder connection(final Channel grpcChannel) {
            this.client = new GatewayClient(grpcChannel);
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
        public GatewayImpl connect() {
            if (this.client == null) {
                throw new IllegalStateException("No gRPC channel has been provided");
            }
            return new GatewayImpl(this);
        }
    }

    private final GatewayClient client;
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
    }

    @Override
    public Identity getIdentity() {
        return this.signingIdentity.getIdentity();
    }

    @Override
    public void close() {
    }

    @Override
    public Network getNetwork(final String networkName) {
        return new NetworkImpl(client, signingIdentity, networkName);
    }
}
