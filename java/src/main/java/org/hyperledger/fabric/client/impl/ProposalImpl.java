/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Proposal;
import org.hyperledger.fabric.client.Transaction;

public class ProposalImpl implements Proposal {
    private final Contract contract;
    private final String transactionName;
    private final List<byte[]> arguments = new ArrayList<>();
    private Map<String, byte[]> transientData;

    ProposalImpl(Contract contract, String transactionName) {
        this.contract = contract;
        this.transactionName = transactionName;
    }

    private String getQualifiedTransactionName() {
        return contract.getContractName()
                .map(contractName -> contractName + ":" + transactionName)
                .orElse(transactionName);
    }

    @Override
    public String getTransactionId() {
        return null;
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
    public Proposal addArguments(final byte[]... args) {
        arguments.addAll(Arrays.asList(args));
        return this;
    }

    @Override
    public Proposal addArguments(final String... args) {
        List<byte[]> byteArgs = Arrays.stream(args)
                .map(arg -> arg.getBytes(StandardCharsets.UTF_8))
                .collect(Collectors.toList());
        arguments.addAll(byteArgs);
        return this;
    }

    @Override
    public Proposal setTransient(final Map<String, byte[]> transientData) {
        this.transientData = new HashMap<>(transientData);
        return this;
    }

    @Override
    public byte[] evaluate() {
        return new byte[0];
    }

    @Override
    public Transaction endorse() {
        return new TransactionImpl();
    }
}
