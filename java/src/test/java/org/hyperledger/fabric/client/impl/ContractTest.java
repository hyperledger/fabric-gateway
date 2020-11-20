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

public class ContractTest {
    private Network network;

    @BeforeEach
    public void beforeEach() throws Exception {
        Gateway gateway = TestUtils.getInstance().newGatewayBuilder().connect();
        network = gateway.getNetwork("ch1");
    }
}
