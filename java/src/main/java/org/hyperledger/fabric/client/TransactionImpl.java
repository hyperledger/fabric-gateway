/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.gateway.CommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;

final class TransactionImpl implements Transaction {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private PreparedTransaction preparedTransaction;

    TransactionImpl(final GatewayClient client, final SigningIdentity signingIdentity,
            final String channelName, final PreparedTransaction preparedTransaction) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.preparedTransaction = preparedTransaction;
    }

    @Override
    public byte[] getResult() {
        return preparedTransaction.getResult()
                .getPayload()
                .toByteArray();
    }

    @Override
    public byte[] getBytes() {
        return preparedTransaction.toByteArray();
    }

    @Override
    public byte[] getDigest() {
        byte[] message = preparedTransaction.getEnvelope().getPayload().toByteArray();
        return signingIdentity.hash(message);
    }

    @Override
    public String getTransactionId() {
        return preparedTransaction.getTransactionId();
    }

    @Override
    public SubmittedTransaction submitAsync() {
        sign();
        SubmitRequest submitRequest = SubmitRequest.newBuilder()
                .setTransactionId(preparedTransaction.getTransactionId())
                .setChannelId(channelName)
                .setPreparedTransaction(preparedTransaction.getEnvelope())
                .build();
        client.submit(submitRequest);

        return new SubmittedTransactionImpl(client, signingIdentity, getTransactionId(), newSignedCommitStatusRequest(), getResult());
    }

    @Override
    public byte[] submit() throws CommitException {
        Status status = submitAsync().getStatus();
        if (!status.isSuccessful()) {
            throw new CommitException(status);
        }

        return getResult();
    }

    void setSignature(final byte[] signature) {
        Common.Envelope envelope = preparedTransaction.getEnvelope().toBuilder()
                .setSignature(ByteString.copyFrom(signature))
                .build();

        preparedTransaction = preparedTransaction.toBuilder()
                .setEnvelope(envelope)
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
        return !preparedTransaction.getEnvelope().getSignature().isEmpty();
    }

    private SignedCommitStatusRequest newSignedCommitStatusRequest() {
        return SignedCommitStatusRequest.newBuilder()
                .setRequest(newCommitStatusRequest().toByteString())
                .build();
    }

    private CommitStatusRequest newCommitStatusRequest() {
        ByteString creator = ByteString.copyFrom(signingIdentity.getCreator());
        return CommitStatusRequest.newBuilder()
                .setChannelId(channelName)
                .setTransactionId(preparedTransaction.getTransactionId())
                .setIdentity(creator)
                .build();
    }
}
