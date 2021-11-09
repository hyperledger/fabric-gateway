/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;

final class NetworkImpl implements Network {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final String channelName;

    NetworkImpl(final GatewayClient client, final SigningIdentity signingIdentity, final String channelName) {
        Objects.requireNonNull(channelName, "network name");

        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
    }

    @Override
    public Contract getContract(final String chaincodeName, final String contractName) {
        return new ContractImpl(client, signingIdentity, contractName, chaincodeName, contractName);
    }

    @Override
    public Contract getContract(final String chaincodeName) {
        return new ContractImpl(client, signingIdentity, channelName, chaincodeName);
    }

    @Override
    public String getName() {
        return channelName;
    }

    @Override
    public CloseableIterator<ChaincodeEvent> getChaincodeEvents(final String chaincodeName, final CallOption... options) {
        return newChaincodeEventsRequest(chaincodeName).build().getEvents(options);
    }

    @Override
    public ChaincodeEventsRequest.Builder newChaincodeEventsRequest(final String chaincodeName) {
        return new ChaincodeEventsBuilder(client, signingIdentity, channelName, chaincodeName);
    }
}
