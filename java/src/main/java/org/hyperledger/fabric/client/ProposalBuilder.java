/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

import com.google.protobuf.ByteString;
import com.google.protobuf.Timestamp;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.gateway.ProposedTransaction;
import org.hyperledger.fabric.protos.peer.Chaincode;
import org.hyperledger.fabric.protos.peer.ProposalPackage;

final class ProposalBuilder implements Proposal.Builder {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private final Chaincode.ChaincodeID chaincodeId;
    private final Chaincode.ChaincodeInput.Builder inputBuilder = Chaincode.ChaincodeInput.newBuilder();
    private final ProposalPackage.ChaincodeProposalPayload.Builder payloadBuilder = ProposalPackage.ChaincodeProposalPayload.newBuilder();
    private Set<String> endorsingOrgs = Collections.emptySet();

    ProposalBuilder(final GatewayClient client, final SigningIdentity signingIdentity,
                    final String channelName, final String chaincodeName, final String transactionName) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.chaincodeId = Chaincode.ChaincodeID.newBuilder()
                .setName(chaincodeName)
                .build();

        inputBuilder.addArgs(ByteString.copyFrom(transactionName, StandardCharsets.UTF_8));
    }

    @Override
    public ProposalBuilder addArguments(final byte[]... args) {
        for (byte[] arg : args) {
            inputBuilder.addArgs(ByteString.copyFrom(arg));
        }
        return this;
    }

    @Override
    public ProposalBuilder addArguments(final String... args) {
        for (String arg : args) {
            inputBuilder.addArgs(ByteString.copyFrom(arg, StandardCharsets.UTF_8));
        }
        return this;
    }

    @Override
    public ProposalBuilder putAllTransient(final Map<String, byte[]> transientData) {
        transientData.forEach(this::putTransient);
        return this;
    }

    @Override
    public ProposalBuilder putTransient(final String key, final byte[] value) {
        payloadBuilder.putTransientMap(key, ByteString.copyFrom(value));
        return this;
    }

    @Override
    public ProposalBuilder putTransient(final String key, final String value) {
        payloadBuilder.putTransientMap(key, ByteString.copyFromUtf8(value));
        return this;
    }

    @Override
    public ProposalBuilder setEndorsingOrganizations(final String... mspids) {
        this.endorsingOrgs = new HashSet<>(Arrays.asList(mspids));
        return this;
    }

    @Override
    public Proposal build() {
        return new ProposalImpl(client, signingIdentity, channelName, newProposedTransaction());
    }

    private ProposedTransaction newProposedTransaction() {
        TransactionContext context = new TransactionContext(signingIdentity);

        return ProposedTransaction.newBuilder()
                .setProposal(newSignedProposal(context))
                .setTransactionId(context.getTransactionId())
                .addAllEndorsingOrganizations(endorsingOrgs)
                .build();
    }

    private ProposalPackage.SignedProposal newSignedProposal(final TransactionContext context) {
        return ProposalPackage.SignedProposal.newBuilder()
                .setProposalBytes(newProposal(context).toByteString())
                .build();
    }

    private ProposalPackage.Proposal newProposal(final TransactionContext context) {
        return ProposalPackage.Proposal.newBuilder()
                .setHeader(newHeader(context).toByteString())
                .setPayload(newChaincodeProposalPayload().toByteString())
                .build();
    }

    private Common.Header newHeader(final TransactionContext context) {
        return Common.Header.newBuilder()
                .setChannelHeader(newChannelHeader(context).toByteString())
                .setSignatureHeader(context.getSignatureHeader().toByteString())
                .build();
    }

    private Common.ChannelHeader newChannelHeader(final TransactionContext context) {
        Timestamp timestamp = Timestamp.newBuilder()
                .setSeconds(Instant.now().getEpochSecond())
                .build();

        return Common.ChannelHeader.newBuilder()
                .setType(Common.HeaderType.ENDORSER_TRANSACTION.getNumber())
                .setTxId(context.getTransactionId())
                .setTimestamp(timestamp)
                .setChannelId(channelName)
                .setExtension(newChaincodeHeaderExtension().toByteString())
                .setEpoch(0)
                .build();
    }

    private ProposalPackage.ChaincodeHeaderExtension newChaincodeHeaderExtension() {
        return ProposalPackage.ChaincodeHeaderExtension.newBuilder()
                .setChaincodeId(chaincodeId)
                .build();
    }

    private ProposalPackage.ChaincodeProposalPayload newChaincodeProposalPayload() {
        return payloadBuilder
                .setInput(newChaincodeInvocationSpec().toByteString())
                .build();
    }

    private Chaincode.ChaincodeInvocationSpec newChaincodeInvocationSpec() {
        return Chaincode.ChaincodeInvocationSpec.newBuilder()
                .setChaincodeSpec(newChaincodeSpec())
                .build();
    }

    private Chaincode.ChaincodeSpec newChaincodeSpec() {
        return Chaincode.ChaincodeSpec.newBuilder()
                .setChaincodeId(chaincodeId)
                .setInput(inputBuilder.build())
                .build();
    }
}
