/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.Supplier;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.gateway.CommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.hyperledger.fabric.protos.peer.TransactionPackage;

class TransactionImpl implements Transaction {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private PreparedTransaction preparedTransaction;

    TransactionImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity,
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
    public Supplier<TransactionPackage.TxValidationCode> submitAsync() {
        submit();

        return () -> {
            CommitStatusRequest statusRequest = CommitStatusRequest.newBuilder()
                    .setChannelId(channelName)
                    .setTransactionId(preparedTransaction.getTransactionId())
                    .build();
            return client.commitStatus(statusRequest).getResult();
        };
    }

    @Override
    public byte[] submitSync() throws ContractException {
        TransactionPackage.TxValidationCode status = submitAsync().get();
        if (!TransactionPackage.TxValidationCode.VALID.equals(status)) {
            throw new ContractException("Commit of transaction " + getTransactionId() + " failed with status code "
                    + status.getNumber() + " (" + status.name() + ")");
        }

        return getResult();
    }

    private void submit() {
        sign();
        SubmitRequest submitRequest = SubmitRequest.newBuilder()
                .setTransactionId(preparedTransaction.getTransactionId())
                .setChannelId(channelName)
                .setPreparedTransaction(preparedTransaction.getEnvelope())
                .build();
        client.submit(submitRequest);
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
}
