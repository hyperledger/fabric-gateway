/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.security.GeneralSecurityException;
import java.security.interfaces.ECPrivateKey;

import io.grpc.Channel;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.Signers;
import org.hyperledger.fabric.client.identity.X509Credentials;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

public class GatewayTest {
    private static final X509Credentials credentials = new X509Credentials();
    private static final Identity identity = new X509Identity("MSP_ID", credentials.getCertificate());
    private static final Signer signer = Signers.newPrivateKeySigner((ECPrivateKey) credentials.getPrivateKey());

    private Gateway gateway;

    @AfterEach
    void afterEach() {
        if (gateway != null) {
            gateway.close();
        }
    }

    @Test
    void connect_with_no_identity_throws() {
        Gateway.Builder builder = Gateway.createBuilder()
                .endpoint("localhost:1337");

        assertThatThrownBy(() -> builder.connect())
                .isInstanceOf(IllegalStateException.class);
    }

    @Test
    void connect_with_no_connection_details_throws() {
        Gateway.Builder builder = Gateway.createBuilder()
                .identity(identity)
                .signer(signer);

        assertThatThrownBy(() -> builder.connect())
                .isInstanceOf(IllegalStateException.class);
    }

    @Test
    void uses_supplied_identity() {
        gateway = Gateway.createBuilder()
                .identity(identity)
                .endpoint("localhost:1337")
                .connect();

        Identity result = ((GatewayImpl) gateway).getIdentity();

        assertThat(result).isEqualTo(identity);
    }

    @Test
    void uses_supplied_signer() throws GeneralSecurityException {
        byte[] expected = "RESULT".getBytes();
        gateway = Gateway.createBuilder()
                .identity(identity)
                .signer((byte[] digest) -> expected)
                .endpoint("localhost:1337")
                .connect();

        byte[] actual = ((GatewayImpl) gateway).sign("DIGEST".getBytes());

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void uses_invalid_signer_if_none_supplied() {
        gateway = Gateway.createBuilder()
                .identity(identity)
                .endpoint("localhost:1337")
                .connect();

        assertThatThrownBy(() -> ((GatewayImpl) gateway).sign("DIGEST".getBytes()))
                .isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void can_connect_using_gRPC_channel() {
        Channel channel = ManagedChannelBuilder.forAddress("localhost", 1337).usePlaintext().build();

        gateway = Gateway.createBuilder()
                .identity(identity)
                .signer(signer)
                .connection(channel)
                .connect();

        assertThat(gateway).isNotNull();
    }

    @Test
    void close_does_not_shutdown_supplied_gRPC_channel() {
        ManagedChannel channel = ManagedChannelBuilder.forAddress("localhost", 1337).usePlaintext().build();
        gateway = Gateway.createBuilder()
                .identity(identity)
                .signer(signer)
                .connection(channel)
                .connect();

        gateway.close();

        assertThat(channel.isShutdown()).isFalse();
    }

    @Test
    void getNetwork_returns_correctly_named_network() {
        gateway = Gateway.createBuilder()
                .identity(identity)
                .signer(signer)
                .endpoint("localhost:1337")
                .connect();

        Network network = gateway.getNetwork("CHANNEL_NAME");

        assertThat(network.getName()).isEqualTo("CHANNEL_NAME");
    }
}
