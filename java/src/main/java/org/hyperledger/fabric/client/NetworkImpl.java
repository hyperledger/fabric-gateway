/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;
import java.util.function.UnaryOperator;

import io.grpc.CallOptions;
import org.hyperledger.fabric.protos.common.Block;
import org.hyperledger.fabric.protos.peer.BlockAndPrivateData;
import org.hyperledger.fabric.protos.peer.FilteredBlock;

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
        return new ContractImpl(client, signingIdentity, channelName, chaincodeName, contractName);
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
    public CloseableIterator<ChaincodeEvent> getChaincodeEvents(final String chaincodeName, final UnaryOperator<CallOptions> options) {
        return newChaincodeEventsRequest(chaincodeName).build().getEvents(options);
    }

    @Override
    public ChaincodeEventsRequest.Builder newChaincodeEventsRequest(final String chaincodeName) {
        return new ChaincodeEventsBuilder(client, signingIdentity, channelName, chaincodeName);
    }

    @Override
    public CloseableIterator<Block> getBlockEvents(final UnaryOperator<CallOptions> options) {
        return newBlockEventsRequest().build().getEvents(options);
    }

    @Override
    public BlockEventsRequest.Builder newBlockEventsRequest() {
        return new BlockEventsBuilder(client, signingIdentity, channelName);
    }

    @Override
    public CloseableIterator<FilteredBlock> getFilteredBlockEvents(final UnaryOperator<CallOptions> options) {
        return newFilteredBlockEventsRequest().build().getEvents(options);
    }

    @Override
    public FilteredBlockEventsRequest.Builder newFilteredBlockEventsRequest() {
        return new FilteredBlockEventsBuilder(client, signingIdentity, channelName);
    }

    @Override
    public CloseableIterator<BlockAndPrivateData> getBlockAndPrivateDataEvents(final UnaryOperator<CallOptions> options) {
        return newBlockAndPrivateDataEventsRequest().build().getEvents(options);
    }

    @Override
    public BlockAndPrivateDataEventsRequest.Builder newBlockAndPrivateDataEventsRequest() {
        return new BlockAndPrivateDataEventsBuilder(client, signingIdentity, channelName);
    }
}
