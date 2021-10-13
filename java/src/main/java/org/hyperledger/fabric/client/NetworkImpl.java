/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;

import com.google.protobuf.InvalidProtocolBufferException;
import org.hyperledger.fabric.protos.gateway.CommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;

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
        return new ContractImpl(client, signingIdentity, contractName, chaincodeName, contractName);
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
    public Commit newSignedCommit(final byte[] bytes, final byte[] signature) throws InvalidProtocolBufferException {
        SignedCommitStatusRequest signedRequest = SignedCommitStatusRequest.parseFrom(bytes);
        CommitStatusRequest request = CommitStatusRequest.parseFrom(signedRequest.getRequest());

        CommitImpl commit = new CommitImpl(client, signingIdentity, request.getTransactionId(), signedRequest);
        commit.setSignature(signature);
        return commit;
    }

    @Override
    public CloseableIterator<ChaincodeEvent> getChaincodeEvents(final String chaincodeName) {
        return newChaincodeEventsRequest(chaincodeName).build().getEvents();
    }

    @Override
    public ChaincodeEventsRequest.Builder newChaincodeEventsRequest(final String chaincodeName) {
        return new ChaincodeEventsBuilder(client, signingIdentity, channelName, chaincodeName);
    }

    @Override
    public ChaincodeEventsRequest newSignedChaincodeEventsRequest(final byte[] bytes, final byte[] signature) throws InvalidProtocolBufferException {
        SignedChaincodeEventsRequest signedRequest = SignedChaincodeEventsRequest.parseFrom(bytes);

        ChaincodeEventsRequestImpl result = new ChaincodeEventsRequestImpl(client, signingIdentity, signedRequest);
        result.setSignature(signature);

        return result;
    }
}
