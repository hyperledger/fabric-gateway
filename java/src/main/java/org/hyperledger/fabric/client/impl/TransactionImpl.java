/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.security.GeneralSecurityException;
import java.util.Iterator;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.client.*;
import org.hyperledger.fabric.gateway.Event;
import org.hyperledger.fabric.gateway.GatewayGrpc;
import org.hyperledger.fabric.gateway.PreparedTransaction;
import org.hyperledger.fabric.gateway.Result;
import org.hyperledger.fabric.protos.common.Common;

public class TransactionImpl implements Transaction {
    private GatewayGrpc.GatewayBlockingStub gatewayService;
    private GatewayImpl gateway;
    private PreparedTransaction preparedTransaction;

    TransactionImpl(GatewayImpl gateway, PreparedTransaction preparedTransaction) {
        this.gateway = gateway;
        this.preparedTransaction = preparedTransaction;
    }

    @Override
    public byte[] getResult() {
        Result result = preparedTransaction.getResponse();
        if(result != null && result.getValue() != null) {
            return result.getValue().toByteArray();
        }
        return new byte[0];
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
        return gateway.getStub().submit(transaction);
    }

    private PreparedTransaction toPreparedTransaction() {
        try {
            // sign the payload
            Common.Envelope envelope = preparedTransaction.getEnvelope();
            byte[] hash = Hash.sha256(envelope.getPayload().toByteArray());
            byte[] signature = gateway.getSigner().sign(hash);
            PreparedTransaction signedTransaction = PreparedTransaction.newBuilder(preparedTransaction)
                    .setEnvelope(Common.Envelope.newBuilder(envelope).setSignature(ByteString.copyFrom(signature)).build())
                    .build();
            return signedTransaction;
        } catch (GeneralSecurityException e) {
            throw new GatewayRuntimeException(e);
        }
    }
}
