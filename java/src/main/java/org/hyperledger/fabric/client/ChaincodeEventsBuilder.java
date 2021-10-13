/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.orderer.Ab;

final class ChaincodeEventsBuilder implements ChaincodeEventsRequest.Builder {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private final String chaincodeName;
    private final Ab.SeekPosition.Builder startPositionBuilder = Ab.SeekPosition.newBuilder()
            .setNextCommit(Ab.SeekNextCommit.getDefaultInstance());

    ChaincodeEventsBuilder(final GatewayClient client, final SigningIdentity signingIdentity, final String channelName,
                           final String chaincodeName) {
        Objects.requireNonNull(channelName, "channel name");
        Objects.requireNonNull(chaincodeName, "chaincode name");

        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.chaincodeName = chaincodeName;
    }

    @Override
    public ChaincodeEventsRequest.Builder startBlock(final long blockNumber) {
        Ab.SeekSpecified specified = Ab.SeekSpecified.newBuilder().setNumber(blockNumber).build();
        startPositionBuilder.setSpecified(specified);
        return this;
    }

    @Override
    public ChaincodeEventsRequest build() {
        SignedChaincodeEventsRequest signedRequest = newSignedChaincodeEventsRequestProto(chaincodeName);
        return new ChaincodeEventsRequestImpl(client, signingIdentity, signedRequest);
    }

    private SignedChaincodeEventsRequest newSignedChaincodeEventsRequestProto(final String chaincodeName) {
        org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest request = newChaincodeEventsRequestProto(chaincodeName);
        return SignedChaincodeEventsRequest.newBuilder()
                .setRequest(request.toByteString())
                .build();
    }

    private org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest newChaincodeEventsRequestProto(final String chaincodeName) {
        ByteString creator = ByteString.copyFrom(signingIdentity.getCreator());
        return org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest.newBuilder()
                .setChannelId(channelName)
                .setChaincodeId(chaincodeName)
                .setIdentity(creator)
                .setStartPosition(startPositionBuilder.build())
                .build();
    }
}
