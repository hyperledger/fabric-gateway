/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.util.concurrent.TimeoutException;

import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.ContractException;
import org.hyperledger.fabric.client.Transaction;

public final class ContractImpl implements Contract {
    private final NetworkImpl network;
    private final String chaincodeId;
    private final String name;

    ContractImpl(final NetworkImpl network, final String chaincodeId, final String name) {
        this.network = network;
        this.chaincodeId = chaincodeId;
        this.name = name;
    }

    @Override
    public Transaction createTransaction(final String name) {
        if (name == null || name.isEmpty()) {
            throw new IllegalArgumentException("Transaction must be a non-empty string");
        }
        String qualifiedName = getQualifiedName(name);
        return new TransactionImpl(this, qualifiedName);
    }

    @Override
    public byte[] submitTransaction(final String name, final String... args) throws ContractException, TimeoutException, InterruptedException {
        return createTransaction(name).submit(args);
    }

    @Override
    public byte[] evaluateTransaction(final String name, final String... args) throws ContractException {
        return createTransaction(name).evaluate(args);
    }

    public NetworkImpl getNetwork() {
        return network;
    }

    public String getChaincodeId() {
        return chaincodeId;
    }

    private String getQualifiedName(final String tname) {
        return this.name.isEmpty() ? tname : this.name + ':' + tname;
    }

}
