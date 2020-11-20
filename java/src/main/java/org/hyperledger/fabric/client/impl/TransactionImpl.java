/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.util.Map;
import java.util.concurrent.TimeoutException;

import org.hyperledger.fabric.client.ContractException;
import org.hyperledger.fabric.client.Transaction;

public final class TransactionImpl implements Transaction {
    private final ContractImpl contract;
    private final String name;
    private final NetworkImpl network = null;
    private final GatewayImpl gateway = null;
    private Map<String, byte[]> transientData = null;

    TransactionImpl(final ContractImpl contract, final String name) {
        this.contract = contract;
        this.name = name;
    }

    @Override
    public String getName() {
        return name;
    }

    @Override
    public Transaction setTransient(final Map<String, byte[]> transientData) {
        this.transientData = transientData;
        return this;
    }

    @Override
    public byte[] submit(final String... args) throws ContractException, TimeoutException, InterruptedException {
        return "".getBytes();
    }

    @Override
    public byte[] evaluate(final String... args) throws ContractException {
        return "".getBytes();
    }
}
