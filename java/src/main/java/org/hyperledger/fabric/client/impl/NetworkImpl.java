/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;


import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Network;

public final class NetworkImpl implements Network {
    private final GatewayImpl gateway;
    private final String name;

    NetworkImpl(final String name, final GatewayImpl gateway) {
        this.gateway = gateway;
        this.name = name;
    }

    @Override
    public Contract getContract(final String chaincodeId, final String name) {
        return new ContractImpl(this, chaincodeId, name);
    }

    @Override
    public Contract getContract(final String chaincodeId) {
        return new ContractImpl(this, chaincodeId);
    }

    @Override
    public GatewayImpl getGateway() {
        return gateway;
    }

    @Override
    public String getName() {
        return name;
    }
}
