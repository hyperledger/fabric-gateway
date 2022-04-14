/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;

import org.hyperledger.fabric.protos.common.Common;

final class BlockAndPrivateDataEventsBuilder implements BlockAndPrivateDataEventsRequest.Builder {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final BlockEventsEnvelopeBuilder envelopeBuilder;

    BlockAndPrivateDataEventsBuilder(final GatewayClient client, final SigningIdentity signingIdentity, final String channelName) {
        Objects.requireNonNull(channelName, "channel name");

        this.client = client;
        this.signingIdentity = signingIdentity;
        this.envelopeBuilder = new BlockEventsEnvelopeBuilder(signingIdentity, channelName);
    }

    @Override
    public BlockAndPrivateDataEventsBuilder startBlock(final long blockNumber) {
        envelopeBuilder.startBlock(blockNumber);
        return this;
    }

    @Override
    public BlockAndPrivateDataEventsBuilder checkpoint(final Checkpoint checkpoint) {
        long blockNumber = checkpoint.getBlockNumber();
        if (blockNumber == 0) {
            return this;
        }
        envelopeBuilder.startBlock(blockNumber);
        return this;
    }

    @Override
    public BlockAndPrivateDataEventsRequest build() {
        Common.Envelope request = envelopeBuilder.build();
        return new BlockAndPrivateDataEventsRequestImpl(client, signingIdentity, request);
    }
}
