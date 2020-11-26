/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import org.hyperledger.fabric.client.*;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

public class ContractTest {
    private Gateway gateway;
    private Network network;

    @BeforeEach
    public void beforeEach() throws Exception {
        gateway = TestUtils.getInstance().newGatewayBuilder().connect();
        network = gateway.getNetwork("ch1");
    }

    @AfterEach
    public void afterEach() {
        gateway.close();
    }

    @Test
    public void evaluateTransaction() throws Exception {
        beforeEach();
        Contract contract = network.getContract("mychannel");
        byte[] result = contract.evaluateTransaction("mytx1", "arg1", "arg2");
        assertThat(result).asString().isEqualTo("mychannelmytx1arg1arg2");
        afterEach();
    }

    @Test
    public void submitTransaction() throws Exception {
        beforeEach();
        Contract contract = network.getContract("channel2");
        byte[] result = contract.submitTransaction("mytx2", "arg");
        assertThat(result).asString().isEqualTo("channel2mytx2arg");
        afterEach();
    }
}
