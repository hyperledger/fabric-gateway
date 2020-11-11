/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.gateway.impl;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import java.io.IOException;

import org.hyperledger.fabric.gateway.Gateway;
import org.hyperledger.fabric.gateway.Identities;
import org.hyperledger.fabric.gateway.Identity;
import org.hyperledger.fabric.gateway.X509Credentials;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class GatewayBuilderTest {
    private static final String GATEWAY_URL = "localhost:7053";

    private final X509Credentials credentials = new X509Credentials();
    private final Identity identity = Identities.newX509Identity("msp1", credentials.getCertificate(), credentials.getPrivateKey());
    private Gateway.Builder builder;

    @BeforeEach
    public void setup() throws IOException {
        builder = Gateway.createBuilder();
    }

    // @Test
    // public void testBuilderNoOptions() {
    //     assertThatThrownBy(() -> builder.connect())
    //             .isInstanceOf(IllegalStateException.class)
    //             .hasMessage("The gateway identity must be set");
    // }

    // @Test
    // public void testBuilderNoCcp() throws IOException {
    //     builder.identity(identity);
    //     assertThatThrownBy(() -> builder.connect())
    //             .isInstanceOf(IllegalStateException.class)
    //             .hasMessage("The network configuration must be specified");
    // }

    // @Test
    // public void testBuilderInvalidIdentity() {
    //     assertThatThrownBy(() -> builder.identity(null))
    //             .isInstanceOf(IllegalArgumentException.class)
    //             .hasMessageContaining("INVALID_IDENTITY");
    // }

    @Test
    public void testBuilderForUnsupportedIdentityType() throws IOException {
        Identity unsupportedIdentity = new Identity() {
            @Override
            public String getMspId() {
                return "mspId";
            }
        };
        assertThatThrownBy(() -> builder.identity(unsupportedIdentity))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining(unsupportedIdentity.getClass().getName());
    }

    @Test
    public void testBuilderWithoutIdentity() throws IOException {
        assertThatThrownBy(() -> builder.identity(null))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Identity must not be null");
    }

    @Test
    public void testBuilderWithIdentity() throws IOException {
        builder.identity(identity)
                .networkConfig(GATEWAY_URL);
        try (Gateway gateway = builder.connect()) {
            assertThat(gateway.getIdentity()).isEqualTo(identity);
        }
    }

}
