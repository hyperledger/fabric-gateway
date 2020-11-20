/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.util.function.Function;

import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;

public final class GatewayImpl implements Gateway {
    public static final class Builder implements Gateway.Builder {
        private String url;
        private Identity identity;
        private Signer signer;

        @Override
        public Builder endpoint(final String url) {
            this.url = url;
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

    private final String url;
    private final Identity identity;
    private final Signer signer;
    private final Function<byte[], byte[]> hash = Hash::sha256;

    private GatewayImpl(final Builder builder) {
        this.url = builder.url;
        this.identity = builder.identity;
        // No signer implementation is required if only offline signing is used
        this.signer = builder.signer != null ? builder.signer : (byte[] digest) -> {
            throw new IllegalStateException("No signing implementation supplied");
        };

        if (null == this.identity) {
            throw new IllegalStateException("No client identity supplied");
        }
    }

    @Override
    public void close() {
        // no op
    }

    @Override
    public Network getNetwork(final String networkName) {
        return new NetworkImpl(networkName, this);
    }

    @Override
    public Identity getIdentity() {
        return identity;
    }

    @Override
    public Signer getSigner() {
        return signer;
    }

    public byte[] hash(byte[] message) {
        return hash(message);
    }
}
