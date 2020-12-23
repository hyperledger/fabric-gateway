/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Map;

import com.google.protobuf.ByteString;
import com.google.protobuf.Timestamp;
import org.hyperledger.fabric.gateway.GatewayGrpc;
import org.hyperledger.fabric.gateway.PreparedTransaction;
import org.hyperledger.fabric.gateway.ProposedTransaction;
import org.hyperledger.fabric.gateway.Result;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.peer.Chaincode;
import org.hyperledger.fabric.protos.peer.ProposalPackage;

class ProposalImpl implements Proposal {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private final String chaincodeId;
    private final Chaincode.ChaincodeInput.Builder inputBuilder = Chaincode.ChaincodeInput.newBuilder();
    private final ProposalPackage.ChaincodeProposalPayload.Builder payloadBuilder = ProposalPackage.ChaincodeProposalPayload.newBuilder();
    private final TransactionContext context;

    ProposalImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity,
                 final String channelName, final String chaincodeId, final String transactionName) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.chaincodeId = chaincodeId;

        context = new TransactionContext(signingIdentity);

        inputBuilder.addArgs(ByteString.copyFrom(transactionName, StandardCharsets.UTF_8));
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
    public byte[] getDigest() {
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

        Result result = client.evaluate(proposedTransaction);
        if (result != null && result.getValue() != null) {
            return result.getValue().toByteArray();
        }

        return new byte[0];
    }

    @Override
    public Transaction endorse() {
        ProposedTransaction proposedTransaction = createProposedTransaction();
        PreparedTransaction preparedTransaction = client.endorse(proposedTransaction);
        return new TransactionImpl(client, signingIdentity, preparedTransaction);
    }

    private ProposedTransaction createProposedTransaction() {
        ProposalPackage.Proposal proposal = createProposal();
        ProposalPackage.SignedProposal signedProposal = signProposal(proposal);
        ProposedTransaction proposedTransaction = ProposedTransaction.newBuilder()
                .setProposal(signedProposal)
                .setTxId(getTransactionId())
                .setChannelId(channelName)
                .build();
        return proposedTransaction;
    }

    private ProposalPackage.Proposal createProposal() {
        Chaincode.ChaincodeID chaincodeID = Chaincode.ChaincodeID.newBuilder().setName(chaincodeId).build();
        ProposalPackage.ChaincodeHeaderExtension chaincodeHeaderExtension = ProposalPackage.ChaincodeHeaderExtension.newBuilder()
                .setChaincodeId(chaincodeID)
                .build();
        Common.ChannelHeader channelHeader = Common.ChannelHeader.newBuilder()
                .setType(Common.HeaderType.ENDORSER_TRANSACTION.getNumber())
                .setTxId(getTransactionId())
                .setTimestamp(Timestamp.newBuilder().setSeconds(Instant.now().getEpochSecond()).build())
                .setChannelId(channelName)
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

    private ProposalPackage.SignedProposal signProposal(final ProposalPackage.Proposal proposal) {
        ByteString payload = proposal.toByteString();
        byte[] hash = signingIdentity.hash(payload.toByteArray());
        byte[] signature = signingIdentity.sign(hash);
        ProposalPackage.SignedProposal signedProposal = ProposalPackage.SignedProposal.newBuilder()
                .setProposalBytes(payload)
                .setSignature(ByteString.copyFrom(signature))
                .build();
        return signedProposal;
    }
}
