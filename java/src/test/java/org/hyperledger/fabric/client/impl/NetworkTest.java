/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.TestUtils;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

public class NetworkTest {
    private static final TestUtils testUtils = TestUtils.getInstance();

    private Gateway gateway;
    private Network network;

    @BeforeEach
    public void beforeEach() throws Exception {
        gateway = testUtils.newGatewayBuilder().connect();
        network = gateway.getNetwork("ch1");
    }

    @AfterEach
    public void afterEach() {
        gateway.close();
    }

    @Test
    public void testGetChannel() {
        assertThat(network.getName()).isEqualTo("ch1");
    }

    @Test
    public void testGetGateway() {
        Gateway gw = network.getGateway();
        assertThat(gw).isSameAs(gateway);
    }

    @Test
    public void testGetContract() {
        Contract contract = network.getContract("contract1");
        assertThat(contract).isInstanceOf(ContractImpl.class);
    }

    @Test
    public void testGetContractEmptyId() {
        assertThatThrownBy(() -> network.getContract(""))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("getContract: chaincodeId must be a non-empty string");
    }

    @Test
    public void testGetContractNullId() {
        assertThatThrownBy(() -> network.getContract(null))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("getContract: chaincodeId must be a non-empty string");
    }

    @Test
    public void testGetContractNullName() {
        assertThatThrownBy(() -> network.getContract("id", null))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("getContract: name must not be null");
    }

}
