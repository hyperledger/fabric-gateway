/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.security.GeneralSecurityException;
import java.util.function.Function;

import io.grpc.Channel;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.GatewayRuntimeException;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;

public final class GatewayImpl implements Gateway {
    private static final Signer UNDEFINED_SIGNER = (byte[] digest) -> {
        throw new UnsupportedOperationException("No signing implementation supplied");
    };

    public static final class Builder implements Gateway.Builder {
        private GatewayClient client;
        private Identity identity;
        private Signer signer = UNDEFINED_SIGNER; // No signer implementation is required if only offline signing is used

        @Override
        public Builder endpoint(final String url) { // TODO: Maybe should be abstracted out to Endpoint class
            this.client = GatewayClientImpl.fromEndpoint(url);
            return this;
        }

        @Override
        public Gateway.Builder connection(final Channel grpcChannel) {
            this.client = GatewayClientImpl.fromChannel(grpcChannel);
            return this;
        }

        public Gateway.Builder client(final GatewayClient client) {
            this.client = client;
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

    private final GatewayClient client;
    private final Identity identity;
    private final Signer signer;
    private final Function<byte[], byte[]> hash = Hash::sha256;

    private GatewayImpl(final Builder builder) {
        this.identity = builder.identity;
        this.signer = builder.signer;
        this.client = builder.client;

        if (null == this.identity) {
            throw new IllegalStateException("No client identity supplied");
        }
        if (null == this.client) {
            throw new IllegalStateException("No connections details supplied");
        }
    }

    @Override
    public void close() {
        client.close();
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
        return hash.apply(message);
    }

    public GatewayClient getClient() {
        return client;
    }
}
