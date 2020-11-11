/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.gateway.impl;

import java.io.IOException;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.hyperledger.fabric.gateway.Gateway;
import org.hyperledger.fabric.gateway.Identity;
import org.hyperledger.fabric.gateway.Network;
import org.hyperledger.fabric.gateway.X509Identity;

public final class GatewayImpl implements Gateway {
    private static final Log LOG = LogFactory.getLog(Gateway.class);

    private final String url;
    private final Identity identity;

    public static final class Builder implements Gateway.Builder {
        private String url = null;
        private Identity identity = null;

        @Override
        public Builder networkConfig(final String url) throws IOException {
            this.url = url;
            return this;
        }

        @Override
        public Builder identity(final Identity identity) {
            if (null == identity) {
                throw new IllegalArgumentException("Identity must not be null");
            }
            if (!(identity instanceof X509Identity)) {
                throw new IllegalArgumentException("No provider for identity type: " + identity.getClass().getName());
            }
            this.identity = identity;
            return this;
        }

        @Override
        public GatewayImpl connect() {
            return new GatewayImpl(this);
        }
    }

    private GatewayImpl(final Builder builder) {
        this.url = builder.url;
        this.identity = builder.identity;
    }

    @Override
    public synchronized void close() {
        // no op
    }

    @Override
    public synchronized Network getNetwork(final String networkName) {
        return new NetworkImpl(networkName, this);
    }

    @Override
    public Identity getIdentity() {
        return identity;
    }

    // public GatewayImpl newInstance() {
    //     return new GatewayImpl(this);
    // }

}
