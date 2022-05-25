/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.UnaryOperator;

import com.google.protobuf.ByteString;
import io.grpc.CallOptions;
import org.hyperledger.fabric.protos.gateway.CommitStatusResponse;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;

class CommitImpl implements Commit {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final String transactionId;
    private SignedCommitStatusRequest signedRequest;

    CommitImpl(final GatewayClient client, final SigningIdentity signingIdentity,
                final String transactionId, final SignedCommitStatusRequest signedRequest) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.transactionId = transactionId;
        this.signedRequest = signedRequest;
    }

    @Override
    public byte[] getBytes() {
        return signedRequest.toByteArray();
    }

    @Override
    public byte[] getDigest() {
        byte[] message = signedRequest.getRequest().toByteArray();
        return signingIdentity.hash(message);
    }

    @Override
    public String getTransactionId() {
        return this.transactionId;
    }

    @Override
    public Status getStatus(final UnaryOperator<CallOptions> options) throws CommitStatusException {
        sign();
        CommitStatusResponse response = client.commitStatus(signedRequest, options);
        return new StatusImpl(transactionId, response);
    }

    void setSignature(final byte[] signature) {
        signedRequest = signedRequest.toBuilder()
                .setSignature(ByteString.copyFrom(signature))
                .build();
    }

    private void sign() {
        if (isSigned()) {
            return;
        }

        byte[] digest = getDigest();
        byte[] signature = signingIdentity.sign(digest);
        setSignature(signature);
    }

    private boolean isSigned() {
        return !signedRequest.getSignature().isEmpty();
    }
}
