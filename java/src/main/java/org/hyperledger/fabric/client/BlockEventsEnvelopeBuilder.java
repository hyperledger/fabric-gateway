/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.OptionalLong;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.orderer.Ab;

final class BlockEventsEnvelopeBuilder {
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private final StartPositionBuilder startPositionBuilder = new StartPositionBuilder();

    BlockEventsEnvelopeBuilder(final SigningIdentity signingIdentity, final String channelName) {
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
    }

    public BlockEventsEnvelopeBuilder startBlock(final long blockNumber) {
        startPositionBuilder.startBlock(blockNumber);
        return this;
    }

    public OptionalLong checkpoint(final Checkpoint checkpoint) {
        return startPositionBuilder.checkpoint(checkpoint);
    }

    public Common.Envelope build() {
        return Common.Envelope.newBuilder()
                .setPayload(newPayload().toByteString())
                .build();
    }

    private Common.Payload newPayload() {
        return Common.Payload.newBuilder()
                .setHeader(newHeader())
                .setData(newSeekInfo().toByteString())
                .build();
    }

    private Common.Header newHeader() {
        return Common.Header.newBuilder()
                .setChannelHeader(newChannelHeader().toByteString())
                .setSignatureHeader(newSignatureHeader().toByteString())
                .build();
    }

    private Common.ChannelHeader newChannelHeader() {
        return Common.ChannelHeader.newBuilder()
                .setChannelId(channelName)
                .setEpoch(0)
                .setTimestamp(GatewayUtils.getCurrentTimestamp())
                .setType(Common.HeaderType.DELIVER_SEEK_INFO_VALUE)
                .build();
    }

    private Common.SignatureHeader newSignatureHeader() {
        ByteString creator = ByteString.copyFrom(signingIdentity.getCreator());
        return Common.SignatureHeader.newBuilder()
                .setCreator(creator)
                .build();
    }

    private Ab.SeekInfo newSeekInfo() {
        return Ab.SeekInfo.newBuilder()
                .setStart(startPositionBuilder.build())
                .setStop(GatewayUtils.seekLargestBlockNumber())
                .build();
    }
}
