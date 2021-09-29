/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;
import java.util.Objects;

import com.google.protobuf.InvalidProtocolBufferException;
import org.hyperledger.fabric.protos.gateway.CommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;

final class NetworkImpl implements Network {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private final String channelName;

    NetworkImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity, final String channelName) {
        Objects.requireNonNull(channelName, "network name");

        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
    }

    @Override
    public Contract getContract(final String chaincodeId, final String contractName) {
        return new ContractImpl(client, signingIdentity, contractName, chaincodeId, contractName);
    }

    @Override
    public Contract getContract(final String chaincodeId) {
        return new ContractImpl(client, signingIdentity, channelName, chaincodeId);
    }

    @Override
    public String getName() {
        return channelName;
    }

    @Override
    public Commit newSignedCommit(final byte[] bytes, final byte[] signature) throws InvalidProtocolBufferException {
        SignedCommitStatusRequest signedRequest = SignedCommitStatusRequest.parseFrom(bytes);
        CommitStatusRequest request = CommitStatusRequest.parseFrom(signedRequest.getRequest());

        CommitImpl commit = new CommitImpl(client, signingIdentity, request.getTransactionId(), signedRequest);
        commit.setSignature(signature);
        return commit;
    }

    @Override
    public Iterator<ChaincodeEvent> getChaincodeEvents(final String chaincodeId) {
        return newChaincodeEventsRequest(chaincodeId).build().getEvents();
    }

    @Override
    public ChaincodeEventsRequest.Builder newChaincodeEventsRequest(final String chaincodeId) {
        return new ChaincodeEventsBuilder(client, signingIdentity, channelName, chaincodeId);
    }

    @Override
    public ChaincodeEventsRequest newSignedChaincodeEventsRequest(final byte[] bytes, final byte[] signature) throws InvalidProtocolBufferException {
        SignedChaincodeEventsRequest signedRequest = SignedChaincodeEventsRequest.parseFrom(bytes);

        ChaincodeEventsRequestImpl result = new ChaincodeEventsRequestImpl(client, signingIdentity, signedRequest);
        result.setSignature(signature);

        return result;
    }
}
