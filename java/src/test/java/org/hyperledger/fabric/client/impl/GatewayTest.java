/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.TestUtils;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

public class GatewayTest {
    private static final TestUtils testUtils = TestUtils.getInstance();

    private Gateway.Builder builder = null;

    @BeforeEach
    public void beforeEach() throws Exception {
        builder = testUtils.newGatewayBuilder();
    }

    @Test
    public void getNetwork_returns_correctly_named_network() {
        try (Gateway gateway = builder.connect()) {
            Network network = gateway.getNetwork("mychannel");
            assertThat(network.getName()).isEqualTo("mychannel");
        }
    }

    // @Test
    // public void testGetNetworkEmptyString() {
    //     try (Gateway gateway = builder.connect()) {
    //         assertThatThrownBy(() -> gateway.getNetwork(""))
    //                 .isInstanceOf(IllegalArgumentException.class)
    //                 .hasMessage("Channel name must be a non-empty string");
    //     }
    // }

    // @Test
    // public void testGetNetworkNullString() {
    //     try (Gateway gateway = builder.connect()) {
    //         assertThatThrownBy(() -> gateway.getNetwork(null))
    //                 .isInstanceOf(IllegalArgumentException.class)
    //                 .hasMessage("Channel name must be a non-empty string");
    //     }
    // }

}
