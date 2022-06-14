/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.UnaryOperator;

import com.google.protobuf.ByteString;
import io.grpc.CallOptions;
import org.hyperledger.fabric.protos.common.Envelope;
import org.hyperledger.fabric.protos.gateway.CommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;

final class TransactionImpl implements Transaction {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private PreparedTransaction preparedTransaction;
    private final ByteString result;

    TransactionImpl(final GatewayClient client, final SigningIdentity signingIdentity, final PreparedTransaction preparedTransaction) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.preparedTransaction = preparedTransaction;

        TransactionEnvelopeParser parser = new TransactionEnvelopeParser(preparedTransaction.getEnvelope());
        this.channelName = parser.getChannelName();
        this.result = parser.getResult();
    }

    @Override
    public byte[] getResult() {
        return result.toByteArray();
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
    public byte[] submit(final UnaryOperator<CallOptions> options) throws SubmitException, CommitStatusException, CommitException {
        Status status = submitAsync(options).getStatus(options);
        if (!status.isSuccessful()) {
            throw new CommitException(status);
        }

        return getResult();
    }

    @Override
    public SubmittedTransaction submitAsync(final UnaryOperator<CallOptions> options) throws SubmitException {
        sign();
        SubmitRequest submitRequest = SubmitRequest.newBuilder()
                .setTransactionId(preparedTransaction.getTransactionId())
                .setChannelId(channelName)
                .setPreparedTransaction(preparedTransaction.getEnvelope())
                .build();
        client.submit(submitRequest, options);

        return new SubmittedTransactionImpl(client, signingIdentity, getTransactionId(), newSignedCommitStatusRequest(), getResult());
    }

    void setSignature(final byte[] signature) {
        Envelope envelope = preparedTransaction.getEnvelope().toBuilder()
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
