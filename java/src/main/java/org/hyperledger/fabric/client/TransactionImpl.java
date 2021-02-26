/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.gateway.Event;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.common.Common;

class TransactionImpl implements Transaction {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private PreparedTransaction preparedTransaction;


    TransactionImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity, final PreparedTransaction preparedTransaction) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.preparedTransaction = preparedTransaction;
    }

    @Override
    public byte[] getResult() {
        return preparedTransaction.getResponse()
                .getValue()
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
        Iterator<Event> eventIter = submit();
        final byte[] result = getResult(); // Get result on current thread, not in Future

        return () -> {
            while (eventIter.hasNext()) {
                Event event = eventIter.next();
                //throw new ContractException("Failed to commit: " + event.getValue().toStringUtf8());
            }
            return result;
        };
    }

    @Override
    public byte[] submitSync() throws ContractException {
        return submitAsync().call();
    }

    private final int sleepTime = 2000;

    private Iterator<Event> submit() {
        sign();
        Iterator<Event> stream = client.submit(preparedTransaction);

        //// TODO remove the following once commit notification has been implemented in the gateway
        try {
            Thread.sleep(sleepTime);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
        /////

        return stream;
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
