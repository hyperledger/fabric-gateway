/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Map;

import com.google.protobuf.ByteString;
import com.google.protobuf.Timestamp;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Proposal;
import org.hyperledger.fabric.client.Transaction;
import org.hyperledger.fabric.gateway.PreparedTransaction;
import org.hyperledger.fabric.gateway.ProposedTransaction;
import org.hyperledger.fabric.gateway.Result;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.peer.Chaincode;
import org.hyperledger.fabric.protos.peer.ProposalPackage;

public class ProposalImpl implements Proposal {
    private final Contract contract;
    private final NetworkImpl network;
    private final GatewayImpl gateway;
    private final String transactionName;
    private final Chaincode.ChaincodeInput.Builder inputBuilder = Chaincode.ChaincodeInput.newBuilder();
    private final ProposalPackage.ChaincodeProposalPayload.Builder payloadBuilder = ProposalPackage.ChaincodeProposalPayload.newBuilder();
    private final TransactionContext context;

    ProposalImpl(ContractImpl contract, String transactionName) {
        this.contract = contract;
        this.transactionName = transactionName;

        network = contract.getNetwork();
        gateway = network.getGateway();

        context = new TransactionContext(gateway.getIdentity(), gateway::hash);

        String qualifiedTxName = contract.getContractName()
                .map(contractName -> contractName + ":" + transactionName)
                .orElse(transactionName);
        inputBuilder.addArgs(ByteString.copyFrom(qualifiedTxName, StandardCharsets.UTF_8));
    }

    @Override
    public String getTransactionId() {
        return context.getTransactionId();
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
        for (byte[] arg : args) {
            inputBuilder.addArgs(ByteString.copyFrom(arg));
        }
        return this;
    }

    @Override
    public Proposal addArguments(final String... args) {
        for (String arg : args) {
            inputBuilder.addArgs(ByteString.copyFrom(arg, StandardCharsets.UTF_8));
        }
        return this;
    }

    @Override
    public Proposal putAllTransient(final Map<String, byte[]> transientData) {
        transientData.forEach(this::putTransient);
        return this;
    }

    @Override
    public Proposal putTransient(final String key, final byte[] value) {
        payloadBuilder.putTransientMap(key, ByteString.copyFrom(value));
        return this;
    }

    @Override
    public byte[] evaluate() {
        ProposedTransaction proposedTransaction = createProposedTransaction();

        Result result = gateway.getService().evaluate(proposedTransaction);
        if (result != null && result.getValue() != null) {
            return result.getValue().toByteArray();
        }

        return new byte[0];
    }

    @Override
    public Transaction endorse() {
        ProposedTransaction proposedTransaction = createProposedTransaction();
        PreparedTransaction preparedTransaction = gateway.getService().endorse(proposedTransaction);
        return new TransactionImpl(gateway, preparedTransaction);
    }

    private ProposedTransaction createProposedTransaction() {
        ProposalPackage.Proposal proposal = createProposal();
        ProposalPackage.SignedProposal signedProposal = signProposal(proposal);
        ProposedTransaction proposedTransaction = ProposedTransaction.newBuilder()
                .setProposal(signedProposal)
                .build();
        return proposedTransaction;
    }

    private ProposalPackage.Proposal createProposal() {
        Chaincode.ChaincodeID chaincodeID = Chaincode.ChaincodeID.newBuilder().setName(contract.getChaincodeId()).build();
        ProposalPackage.ChaincodeHeaderExtension chaincodeHeaderExtension = ProposalPackage.ChaincodeHeaderExtension.newBuilder()
                .setChaincodeId(chaincodeID)
                .build();
        Common.ChannelHeader channelHeader = Common.ChannelHeader.newBuilder()
                .setType(Common.HeaderType.ENDORSER_TRANSACTION.getNumber())
                .setTxId(getTransactionId())
                .setTimestamp(Timestamp.newBuilder().setSeconds(Instant.now().getEpochSecond()).build())
                .setChannelId(network.getName())
                .setExtension(chaincodeHeaderExtension.toByteString())
                .setEpoch(0)
                .build();
        Common.Header header = Common.Header.newBuilder()
                .setChannelHeader(channelHeader.toByteString())
                .setSignatureHeader(context.getSignatureHeader().toByteString())
                .build();
        Chaincode.ChaincodeSpec chaincodeSpec = Chaincode.ChaincodeSpec.newBuilder()
                .setType(Chaincode.ChaincodeSpec.Type.NODE)
                .setChaincodeId(chaincodeID)
                .setInput(inputBuilder.build())
                .build();
        Chaincode.ChaincodeInvocationSpec chaincodeInvocationSpec = Chaincode.ChaincodeInvocationSpec.newBuilder()
                .setChaincodeSpec(chaincodeSpec)
                .build();
        payloadBuilder.setInput(chaincodeInvocationSpec.toByteString());
        ProposalPackage.Proposal proposal = ProposalPackage.Proposal.newBuilder()
                .setHeader(header.toByteString())
                .setPayload(payloadBuilder.build().toByteString())
                .build();

        return proposal;
    }

    private ProposalPackage.SignedProposal signProposal(ProposalPackage.Proposal proposal) {
        ByteString payload = proposal.toByteString();
        byte[] hash = gateway.hash(payload.toByteArray());
        byte[] signature = gateway.sign(hash);
        ProposalPackage.SignedProposal signedProposal = ProposalPackage.SignedProposal.newBuilder()
                .setProposalBytes(payload)
                .setSignature(ByteString.copyFrom(signature))
                .build();
        return signedProposal;
    }
}
