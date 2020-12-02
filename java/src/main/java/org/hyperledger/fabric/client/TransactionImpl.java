/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.gateway.Event;
import org.hyperledger.fabric.gateway.GatewayGrpc;
import org.hyperledger.fabric.gateway.PreparedTransaction;
import org.hyperledger.fabric.gateway.Result;
import org.hyperledger.fabric.protos.common.Common;

class TransactionImpl implements Transaction {
    private static final byte[] EMPTY_RESULT = new byte[0];

    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private final PreparedTransaction preparedTransaction;


    TransactionImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity, final PreparedTransaction preparedTransaction) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.preparedTransaction = preparedTransaction;
    }

    @Override
    public byte[] getResult() {
        Result result = preparedTransaction.getResponse();
        if (null == result) {
            return EMPTY_RESULT;
        }

        ByteString value = result.getValue();
        if (null == value) {
            return EMPTY_RESULT;
        }

        return value.toByteArray();
    }

    @Override
    public byte[] getBytes() {
        return new byte[0];
    }

    @Override
    public byte[] getHash() {
        return new byte[0];
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

    private Iterator<Event> submit() {
        PreparedTransaction transaction = toPreparedTransaction();
        return client.submit(transaction);
    }

    private PreparedTransaction toPreparedTransaction() {
        // sign the payload
        Common.Envelope envelope = preparedTransaction.getEnvelope();
        byte[] hash = signingIdentity.hash(envelope.getPayload().toByteArray());
        byte[] signature = signingIdentity.sign(hash);
        PreparedTransaction signedTransaction = PreparedTransaction.newBuilder(preparedTransaction)
                .setEnvelope(Common.Envelope.newBuilder(envelope).setSignature(ByteString.copyFrom(signature)).build())
                .build();
        return signedTransaction;
    }
}
