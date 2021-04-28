/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.concurrent.TimeUnit;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

public final class NetworkTest {
    private static final TestUtils testUtils = TestUtils.getInstance();

    private Gateway gateway;
    private Network network;
    private ManagedChannel channel;

    @BeforeEach
    void beforeEach() throws Exception {
        channel = ManagedChannelBuilder.forAddress("example.org", 1337).usePlaintext().build();
        gateway = testUtils.newGatewayBuilder().connection(channel).connect();
        network = gateway.getNetwork("ch1");
    }

    @AfterEach
    void afterEach() {
        gateway.close();
        GatewayUtils.shutdownChannel(channel, 5, TimeUnit.SECONDS);
    }

    @Test
    void getContract_using_only_chaincode_ID_returns_correctly_named_Contract() {
        Contract contract = network.getContract("CHAINCODE_ID");
        assertThat(contract.getChaincodeId()).isEqualTo("CHAINCODE_ID");
        assertThat(contract.getContractName()).isEmpty();
    }

    @Test
    void getContract_using_contract_name_returns_correctly_named_Contract() {
        Contract contract = network.getContract("CHAINCODE_ID", "CONTRACT");
        assertThat(contract.getChaincodeId()).isEqualTo("CHAINCODE_ID");
        assertThat(contract.getContractName()).get().isEqualTo("CONTRACT");
    }
}
