/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.util.Optional;

import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.ContractException;
import org.hyperledger.fabric.client.Proposal;
import org.hyperledger.fabric.client.Transaction;

public final class ContractImpl implements Contract {
    private final NetworkImpl network;
    private final String chaincodeId;
    private final String name;

    ContractImpl(final NetworkImpl network, final String chaincodeId) {
        this(network, chaincodeId, null);
    }

    ContractImpl(final NetworkImpl network, final String chaincodeId, final String name) {
        this.network = network;
        this.chaincodeId = chaincodeId;
        this.name = name;
    }

    @Override
    public byte[] submitTransaction(final String name, final String... args) throws ContractException {
        return newProposal(name).addArguments(args).endorse().submitSync();
    }

    @Override
    public byte[] evaluateTransaction(final String name, final String... args) {
        return newProposal(name).addArguments(args).evaluate();
    }

    @Override
    public Proposal newProposal(final String transactionName) {
        return new ProposalImpl(this, transactionName);
    }

    @Override
    public Proposal newSignedProposal(final byte[] proposalBytes, final byte[] signature) {
        return null;
    }

    @Override
    public Transaction newSignedTransaction(final byte[] transactionBytes, final byte[] signature) {
        return null;
    }

    public NetworkImpl getNetwork() {
        return network;
    }

    @Override
    public String getChaincodeId() {
        return chaincodeId;
    }

    @Override
    public Optional<String> getContractName() {
        return Optional.ofNullable(name);
    }
}
