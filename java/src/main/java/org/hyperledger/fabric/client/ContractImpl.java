/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;
import java.util.Optional;

final class ContractImpl implements Contract {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private final String chaincodeName;
    private final String contractName;

    ContractImpl(final GatewayClient client, final SigningIdentity signingIdentity, final String channelName, final String chaincodeName) {
        this(client, signingIdentity, channelName, chaincodeName, null);
    }

    ContractImpl(final GatewayClient client, final SigningIdentity signingIdentity,
                 final String channelName, final String chaincodeName, final String contractName) {
        Objects.requireNonNull(chaincodeName, "chaincode name");

        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.chaincodeName = chaincodeName;
        this.contractName = contractName;
    }

    @Override
    public byte[] submitTransaction(final String name) throws EndorseException, CommitException, SubmitException, CommitStatusException {
        return newProposal(name)
                .build()
                .endorse()
                .submit();
    }

    @Override
    public byte[] submitTransaction(final String name, final String... args) throws EndorseException, CommitException, SubmitException, CommitStatusException {
        return newProposal(name)
                .addArguments(args)
                .build()
                .endorse()
                .submit();
    }

    @Override
    public byte[] submitTransaction(final String name, final byte[]... args) throws EndorseException, CommitException, SubmitException, CommitStatusException {
        return newProposal(name)
                .addArguments(args)
                .build()
                .endorse()
                .submit();
    }

    @Override
    public byte[] evaluateTransaction(final String name) throws GatewayException {
        return newProposal(name)
                .build()
                .evaluate();
    }

    @Override
    public byte[] evaluateTransaction(final String name, final String... args) throws GatewayException {
        return newProposal(name)
                .addArguments(args)
                .build()
                .evaluate();
    }

    @Override
    public byte[] evaluateTransaction(final String name, final byte[]... args) throws GatewayException {
        return newProposal(name)
                .addArguments(args)
                .build()
                .evaluate();
    }

    @Override
    public Proposal.Builder newProposal(final String transactionName) {
        String qualifiedTxName = qualifiedTransactionName(transactionName);
        return new ProposalBuilder(client, signingIdentity, channelName, chaincodeName, qualifiedTxName);
    }

    @Override
    public String getChaincodeName() {
        return chaincodeName;
    }

    @Override
    public Optional<String> getContractName() {
        return Optional.ofNullable(contractName);
    }

    private String qualifiedTransactionName(final String name) {
        Objects.requireNonNull(name, "transaction name");
        return getContractName()
                .map(contractName -> contractName + ":" + name)
                .orElse(name);
    }
}
