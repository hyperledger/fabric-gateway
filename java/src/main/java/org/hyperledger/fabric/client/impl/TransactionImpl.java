/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.util.Iterator;

import org.hyperledger.fabric.client.Commit;
import org.hyperledger.fabric.client.ContractException;
import org.hyperledger.fabric.client.Transaction;
import org.hyperledger.fabric.gateway.Event;
import org.hyperledger.fabric.gateway.GatewayGrpc;
import org.hyperledger.fabric.gateway.PreparedTransaction;

public class TransactionImpl implements Transaction {
    private GatewayGrpc.GatewayBlockingStub gatewayService;

    @Override
    public byte[] getResult() {
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
                throw new ContractException("Failed to commit: " + event.getValue().toStringUtf8());
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
        return gatewayService.submit(transaction);
    }

    private PreparedTransaction toPreparedTransaction() {
        return null;
    }
}
