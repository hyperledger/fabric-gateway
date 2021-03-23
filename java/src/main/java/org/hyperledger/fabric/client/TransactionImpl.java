/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;

import com.google.protobuf.ByteString;

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
    public Commit submitAsync() {
        submit();
        final byte[] result = getResult(); // Get result on current thread, not in Future

        return () -> {
            return result;
        };
    }

    @Override
    public byte[] submitSync() throws ContractException {
        return submitAsync().call();
    }

    private final int sleepTime = 2000;

    private void submit() {
        sign();

        SubmitRequest submitRequest = SubmitRequest.newBuilder()
                .setTransactionId(preparedTransaction.getTransactionId())
                .setChannelId(channelName)
                .setPreparedTransaction(preparedTransaction.getEnvelope())
                .build();
        client.submit(submitRequest);

        //// TODO remove the following once commit notification has been implemented in the gateway
        try {
            Thread.sleep(sleepTime);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
        /////

        return;
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
