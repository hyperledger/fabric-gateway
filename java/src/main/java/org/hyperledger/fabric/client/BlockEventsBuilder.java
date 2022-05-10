/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;
import java.util.OptionalLong;

import org.hyperledger.fabric.protos.common.Common;

final class BlockEventsBuilder implements BlockEventsRequest.Builder {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final BlockEventsEnvelopeBuilder envelopeBuilder;

    BlockEventsBuilder(final GatewayClient client, final SigningIdentity signingIdentity, final String channelName) {
        Objects.requireNonNull(channelName, "channel name");

        this.client = client;
        this.signingIdentity = signingIdentity;
        this.envelopeBuilder = new BlockEventsEnvelopeBuilder(signingIdentity, channelName);
    }

    @Override
    public BlockEventsBuilder startBlock(final long blockNumber) {
        envelopeBuilder.startBlock(blockNumber);
        return this;
    }

    @Override
    public BlockEventsBuilder checkpoint(final Checkpoint checkpoint) {
        OptionalLong blockNumber = envelopeBuilder.checkpoint(checkpoint);
        blockNumber.ifPresent(envelopeBuilder::startBlock);
        return this;
    }

    @Override
    public BlockEventsRequest build() {
        Common.Envelope request = envelopeBuilder.build();
        return new BlockEventsRequestImpl(client, signingIdentity, request);
    }
}
