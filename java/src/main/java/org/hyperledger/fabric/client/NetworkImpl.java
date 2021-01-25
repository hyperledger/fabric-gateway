/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.gateway.GatewayGrpc;

final class NetworkImpl implements Network {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private final String channelName;

    NetworkImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity, final String channelName) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
    }

    @Override
    public Contract getContract(final String chaincodeId, final String name) {
        return new ContractImpl(client, signingIdentity, name, chaincodeId, name);
    }

    @Override
    public Contract getContract(final String chaincodeId) {
        return new ContractImpl(client, signingIdentity, channelName, chaincodeId);
    }

    @Override
    public String getName() {
        return channelName;
    }
}
