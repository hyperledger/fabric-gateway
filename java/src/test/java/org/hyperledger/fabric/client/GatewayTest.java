/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.concurrent.TimeUnit;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.Signers;
import org.hyperledger.fabric.client.identity.X509Credentials;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

public final class GatewayTest {
    private static final X509Credentials credentials = new X509Credentials();
    private static final Identity identity = new X509Identity("MSP_ID", credentials.getCertificate());
    private static final Signer signer = Signers.newPrivateKeySigner(credentials.getPrivateKey());
    private static final TestUtils utils = TestUtils.getInstance();

    private Gateway gateway;
    private ManagedChannel channel;

    @BeforeEach()
    void beforeEach() {
        channel = ManagedChannelBuilder.forAddress("example.org", 1337).usePlaintext().build();
    }

    @AfterEach
    void afterEach() {
        if (gateway != null) {
            gateway.close();
        }
        utils.shutdownChannel(channel, 5, TimeUnit.SECONDS);
    }

    @Test
    void connect_with_no_identity_throws() {
        Gateway.Builder builder = Gateway.newInstance()
                .connection(channel);

        assertThatThrownBy(builder::connect)
                .isInstanceOf(IllegalArgumentException.class);
    }

    @Test
    void connect_with_no_connection_details_throws() {
        Gateway.Builder builder = Gateway.newInstance()
                .identity(identity)
                .signer(signer);

        assertThatThrownBy(builder::connect)
                .isInstanceOf(IllegalArgumentException.class);
    }

    @Test
    void uses_supplied_identity() {
        gateway = Gateway.newInstance()
                .identity(identity)
                .connection(channel)
                .connect();

        Identity result = gateway.getIdentity();

        assertThat(result).isEqualTo(identity);
    }

    @Test
    void can_connect_using_gRPC_channel() {
        gateway = Gateway.newInstance()
                .identity(identity)
                .signer(signer)
                .connection(channel)
                .connect();

        assertThat(gateway).isNotNull();
    }

    @Test
    void close_does_not_shutdown_supplied_gRPC_channel() {
        gateway = Gateway.newInstance()
                .identity(identity)
                .signer(signer)
                .connection(channel)
                .connect();

        gateway.close();

        assertThat(channel.isShutdown()).isFalse();
    }

    @Test
    void getNetwork_returns_correctly_named_network() {
        gateway = Gateway.newInstance()
                .identity(identity)
                .signer(signer)
                .connection(channel)
                .connect();

        Network network = gateway.getNetwork("CHANNEL_NAME");

        assertThat(network.getName()).isEqualTo("CHANNEL_NAME");
    }

    @Test
    void getNetwork_throws_NullPointerException_on_null_network_name() {
        gateway = Gateway.newInstance()
                .identity(identity)
                .signer(signer)
                .connection(channel)
                .connect();

        assertThatThrownBy(() -> gateway.getNetwork(null))
                .isInstanceOf(NullPointerException.class)
                .hasMessageContaining("network name");
    }
}
