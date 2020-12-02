/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Optional;

import org.hyperledger.fabric.gateway.GatewayGrpc;

final class ContractImpl implements Contract {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private final String chaincodeId;
    private final String contractName;

    ContractImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity, final String channelName, final String chaincodeId) {
        this(client, signingIdentity, channelName, chaincodeId, null);
    }

    ContractImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity, final String channelName, final String chaincodeId, final String contractName) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.chaincodeId = chaincodeId;
        this.contractName = contractName;
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
        String qualifiedTxName = qualifiedTransactionName(transactionName);
        return new ProposalImpl(client, signingIdentity, channelName, chaincodeId, qualifiedTxName);
    }

    @Override
    public Proposal newSignedProposal(final byte[] proposalBytes, final byte[] signature) {
        return null;
    }

    @Override
    public Transaction newSignedTransaction(final byte[] transactionBytes, final byte[] signature) {
        return null;
    }

    @Override
    public String getChaincodeId() {
        return chaincodeId;
    }

    @Override
    public Optional<String> getContractName() {
        return Optional.ofNullable(contractName);
    }

    private String qualifiedTransactionName(final String name) {
        return getContractName()
                .map(contractName -> contractName + ":" + name)
                .orElse(name);
    }
}
