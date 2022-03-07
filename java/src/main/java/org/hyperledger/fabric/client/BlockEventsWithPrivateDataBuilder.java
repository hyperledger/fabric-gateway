/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;

import org.hyperledger.fabric.protos.common.Common;

final class BlockEventsWithPrivateDataBuilder implements BlockEventsWithPrivateDataRequest.Builder {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final BlockEventsEnvelopeBuilder envelopeBuilder;

    BlockEventsWithPrivateDataBuilder(final GatewayClient client, final SigningIdentity signingIdentity, final String channelName) {
        Objects.requireNonNull(channelName, "channel name");

        this.client = client;
        this.signingIdentity = signingIdentity;
        this.envelopeBuilder = new BlockEventsEnvelopeBuilder(signingIdentity, channelName);
    }

    @Override
    public BlockEventsWithPrivateDataBuilder startBlock(final long blockNumber) {
        envelopeBuilder.startBlock(blockNumber);
        return this;
    }

    @Override
    public BlockEventsWithPrivateDataRequest build() {
        Common.Envelope request = envelopeBuilder.build();
        return new BlockEventsWithPrivateDataRequestImpl(client, signingIdentity, request);
    }
}
