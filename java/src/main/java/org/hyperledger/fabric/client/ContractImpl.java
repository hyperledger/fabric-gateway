/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Optional;

import com.google.protobuf.InvalidProtocolBufferException;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.ProposedTransaction;

final class ContractImpl implements Contract {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private final String chaincodeId;
    private final String contractName;

    ContractImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity, final String channelName, final String chaincodeId) {
        this(client, signingIdentity, channelName, chaincodeId, null);
    }

    ContractImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity,
                 final String channelName, final String chaincodeId, final String contractName) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.chaincodeId = chaincodeId;
        this.contractName = contractName;
    }

    @Override
    public byte[] submitTransaction(final String name) throws CommitException {
        return newProposal(name)
                .build()
                .endorse()
                .submit();
    }

    @Override
    public byte[] submitTransaction(final String name, final String... args) throws CommitException {
        return newProposal(name)
                .addArguments(args)
                .build()
                .endorse()
                .submit();
    }

    @Override
    public byte[] submitTransaction(final String name, final byte[]... args) throws CommitException {
        return newProposal(name)
                .addArguments(args)
                .build()
                .endorse()
                .submit();
    }

    @Override
    public byte[] evaluateTransaction(final String name) {
        return newProposal(name)
                .build()
                .evaluate();
    }

    @Override
    public byte[] evaluateTransaction(final String name, final String... args) {
        return newProposal(name)
                .addArguments(args)
                .build()
                .evaluate();
    }

    @Override
    public byte[] evaluateTransaction(final String name, final byte[]... args) {
        return newProposal(name)
                .addArguments(args)
                .build()
                .evaluate();
    }

    @Override
    public Proposal.Builder newProposal(final String transactionName) {
        String qualifiedTxName = qualifiedTransactionName(transactionName);
        return new ProposalBuilder(client, signingIdentity, channelName, chaincodeId, qualifiedTxName);
    }

    @Override
    public Proposal newSignedProposal(final byte[] proposalBytes, final byte[] signature) throws InvalidProtocolBufferException {
        ProposedTransaction proposedTransaction = ProposedTransaction.parseFrom(proposalBytes);

        ProposalImpl proposal = new ProposalImpl(client, signingIdentity, channelName, proposedTransaction);
        proposal.setSignature(signature);
        return proposal;
    }

    @Override
    public Transaction newSignedTransaction(final byte[] transactionBytes, final byte[] signature) throws InvalidProtocolBufferException {
        PreparedTransaction preparedTransaction = PreparedTransaction.parseFrom(transactionBytes);

        TransactionImpl transaction = new TransactionImpl(client, signingIdentity, channelName, preparedTransaction);
        transaction.setSignature(signature);
        return transaction;
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
