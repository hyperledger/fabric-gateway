/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.gateway.impl;


import org.hyperledger.fabric.gateway.Contract;
import org.hyperledger.fabric.gateway.Network;

public final class NetworkImpl implements Network {
    private final GatewayImpl gateway;
    private final String name;

    NetworkImpl(final String name, final GatewayImpl gateway) {
        this.gateway = gateway;
        this.name = name;
    }

    @Override
    public Contract getContract(final String chaincodeId, final String name) {
        if (chaincodeId == null || chaincodeId.isEmpty()) {
            throw new IllegalArgumentException("getContract: chaincodeId must be a non-empty string");
        }
        if (name == null) {
            throw new IllegalArgumentException("getContract: name must not be null");
        }

        return new ContractImpl(this, chaincodeId, name);
    }

    @Override
    public Contract getContract(final String chaincodeId) {
        return getContract(chaincodeId, "");
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
