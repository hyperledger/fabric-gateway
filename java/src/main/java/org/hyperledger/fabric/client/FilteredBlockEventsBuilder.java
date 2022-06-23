/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;

import org.hyperledger.fabric.protos.common.Envelope;

final class FilteredBlockEventsBuilder implements FilteredBlockEventsRequest.Builder {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final BlockEventsEnvelopeBuilder envelopeBuilder;

    FilteredBlockEventsBuilder(final GatewayClient client, final SigningIdentity signingIdentity, final String channelName) {
        Objects.requireNonNull(channelName, "channel name");

        this.client = client;
        this.signingIdentity = signingIdentity;
        this.envelopeBuilder = new BlockEventsEnvelopeBuilder(signingIdentity, channelName);
    }

    @Override
    public FilteredBlockEventsBuilder startBlock(final long blockNumber) {
        envelopeBuilder.startBlock(blockNumber);
        return this;
    }

    @Override
    public FilteredBlockEventsBuilder checkpoint(final Checkpoint checkpoint) {
        checkpoint.getBlockNumber().ifPresent(envelopeBuilder::startBlock);
        return this;
    }

    @Override
    public FilteredBlockEventsRequest build() {
        Envelope request = envelopeBuilder.build();
        return new FilteredBlockEventsRequestImpl(client, signingIdentity, request);
    }
}
