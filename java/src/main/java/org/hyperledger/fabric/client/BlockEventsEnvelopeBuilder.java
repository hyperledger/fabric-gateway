/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.common.ChannelHeader;
import org.hyperledger.fabric.protos.common.Envelope;
import org.hyperledger.fabric.protos.common.Header;
import org.hyperledger.fabric.protos.common.HeaderType;
import org.hyperledger.fabric.protos.common.Payload;
import org.hyperledger.fabric.protos.common.SignatureHeader;
import org.hyperledger.fabric.protos.orderer.SeekInfo;

final class BlockEventsEnvelopeBuilder {
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private final ByteString tlsCertificateHash;
    private final StartPositionBuilder startPositionBuilder = new StartPositionBuilder();

    BlockEventsEnvelopeBuilder(
        final SigningIdentity signingIdentity,
        final String channelName,
        final ByteString tlsCertificateHash
    ) {
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.tlsCertificateHash = tlsCertificateHash;
    }

    public BlockEventsEnvelopeBuilder startBlock(final long blockNumber) {
        startPositionBuilder.startBlock(blockNumber);
        return this;
    }

    public Envelope build() {
        return Envelope.newBuilder()
                .setPayload(newPayload().toByteString())
                .build();
    }

    private Payload newPayload() {
        return Payload.newBuilder()
                .setHeader(newHeader())
                .setData(newSeekInfo().toByteString())
                .build();
    }

    private Header newHeader() {
        return Header.newBuilder()
                .setChannelHeader(newChannelHeader().toByteString())
                .setSignatureHeader(newSignatureHeader().toByteString())
                .build();
    }

    private ChannelHeader newChannelHeader() {
        return ChannelHeader.newBuilder()
                .setChannelId(channelName)
                .setEpoch(0)
                .setTimestamp(GatewayUtils.getCurrentTimestamp())
                .setType(HeaderType.DELIVER_SEEK_INFO_VALUE)
                .setTlsCertHash(tlsCertificateHash)
                .build();
    }

    private SignatureHeader newSignatureHeader() {
        ByteString creator = ByteString.copyFrom(signingIdentity.getCreator());
        return SignatureHeader.newBuilder()
                .setCreator(creator)
                .build();
    }

    private SeekInfo newSeekInfo() {
        return SeekInfo.newBuilder()
                .setStart(startPositionBuilder.build())
                .setStop(GatewayUtils.seekLargestBlockNumber())
                .build();
    }
}
