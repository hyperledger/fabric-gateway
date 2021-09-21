/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.orderer.Ab;

final class ChaincodeEventsBuilder implements ChaincodeEventsRequest.Builder {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private final String chaincodeId;
    private final Ab.SeekPosition.Builder startPositionBuilder = Ab.SeekPosition.newBuilder()
            .setNextCommit(Ab.SeekNextCommit.getDefaultInstance());

    ChaincodeEventsBuilder(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity, final String channelName,
                           final String chaincodeId) {
        Objects.requireNonNull(channelName, "channel name");
        Objects.requireNonNull(chaincodeId, "chaincode ID");

        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.chaincodeId = chaincodeId;
    }

    @Override
    public ChaincodeEventsRequest.Builder startBlock(final long blockNumber) {
        Ab.SeekSpecified specified = Ab.SeekSpecified.newBuilder().setNumber(blockNumber).build();
        startPositionBuilder.setSpecified(specified);
        return this;
    }

    @Override
    public ChaincodeEventsRequest build() {
        SignedChaincodeEventsRequest signedRequest = newSignedChaincodeEventsRequestProto(chaincodeId);
        return new ChaincodeEventsRequestImpl(client, signingIdentity, signedRequest);
    }

    private SignedChaincodeEventsRequest newSignedChaincodeEventsRequestProto(final String chaincodeId) {
        org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest request = newChaincodeEventsRequestProto(chaincodeId);
        return SignedChaincodeEventsRequest.newBuilder()
                .setRequest(request.toByteString())
                .build();
    }

    private org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest newChaincodeEventsRequestProto(final String chaincodeId) {
        ByteString creator = ByteString.copyFrom(signingIdentity.getCreator());
        return org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest.newBuilder()
                .setChannelId(channelName)
                .setChaincodeId(chaincodeId)
                .setIdentity(creator)
                .setStartPosition(startPositionBuilder.build())
                .build();
    }
}
