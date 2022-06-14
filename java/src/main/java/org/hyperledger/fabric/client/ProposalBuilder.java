/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.common.ChannelHeader;
import org.hyperledger.fabric.protos.common.Header;
import org.hyperledger.fabric.protos.common.HeaderType;
import org.hyperledger.fabric.protos.gateway.ProposedTransaction;
import org.hyperledger.fabric.protos.peer.ChaincodeHeaderExtension;
import org.hyperledger.fabric.protos.peer.ChaincodeID;
import org.hyperledger.fabric.protos.peer.ChaincodeInput;
import org.hyperledger.fabric.protos.peer.ChaincodeInvocationSpec;
import org.hyperledger.fabric.protos.peer.ChaincodeProposalPayload;
import org.hyperledger.fabric.protos.peer.ChaincodeSpec;
import org.hyperledger.fabric.protos.peer.SignedProposal;

final class ProposalBuilder implements Proposal.Builder {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private final ChaincodeID chaincodeId;
    private final ChaincodeInput.Builder inputBuilder = ChaincodeInput.newBuilder();
    private final ChaincodeProposalPayload.Builder payloadBuilder = ChaincodeProposalPayload.newBuilder();
    private Set<String> endorsingOrgs = Collections.emptySet();

    ProposalBuilder(final GatewayClient client, final SigningIdentity signingIdentity,
                    final String channelName, final String chaincodeName, final String transactionName) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.chaincodeId = ChaincodeID.newBuilder()
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

    private SignedProposal newSignedProposal(final TransactionContext context) {
        return SignedProposal.newBuilder()
                .setProposalBytes(newProposal(context).toByteString())
                .build();
    }

    private org.hyperledger.fabric.protos.peer.Proposal newProposal(final TransactionContext context) {
        return org.hyperledger.fabric.protos.peer.Proposal.newBuilder()
                .setHeader(newHeader(context).toByteString())
                .setPayload(newChaincodeProposalPayload().toByteString())
                .build();
    }

    private Header newHeader(final TransactionContext context) {
        return Header.newBuilder()
                .setChannelHeader(newChannelHeader(context).toByteString())
                .setSignatureHeader(context.getSignatureHeader().toByteString())
                .build();
    }

    private ChannelHeader newChannelHeader(final TransactionContext context) {
        return ChannelHeader.newBuilder()
                .setType(HeaderType.ENDORSER_TRANSACTION.getNumber())
                .setTxId(context.getTransactionId())
                .setTimestamp(GatewayUtils.getCurrentTimestamp())
                .setChannelId(channelName)
                .setExtension(newChaincodeHeaderExtension().toByteString())
                .setEpoch(0)
                .build();
    }

    private ChaincodeHeaderExtension newChaincodeHeaderExtension() {
        return ChaincodeHeaderExtension.newBuilder()
                .setChaincodeId(chaincodeId)
                .build();
    }

    private ChaincodeProposalPayload newChaincodeProposalPayload() {
        return payloadBuilder
                .setInput(newChaincodeInvocationSpec().toByteString())
                .build();
    }

    private ChaincodeInvocationSpec newChaincodeInvocationSpec() {
        return ChaincodeInvocationSpec.newBuilder()
                .setChaincodeSpec(newChaincodeSpec())
                .build();
    }

    private ChaincodeSpec newChaincodeSpec() {
        return ChaincodeSpec.newBuilder()
                .setChaincodeId(chaincodeId)
                .setInput(inputBuilder.build())
                .build();
    }
}
