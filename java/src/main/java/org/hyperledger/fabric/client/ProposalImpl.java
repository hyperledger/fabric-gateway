/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.UnaryOperator;

import com.google.protobuf.ByteString;
import io.grpc.CallOptions;
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.EndorseResponse;
import org.hyperledger.fabric.protos.gateway.EvaluateRequest;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.ProposedTransaction;
import org.hyperledger.fabric.protos.peer.SignedProposal;

final class ProposalImpl implements Proposal {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final String channelName;
    private ProposedTransaction proposedTransaction;

    ProposalImpl(final GatewayClient client, final SigningIdentity signingIdentity,
            final String channelName, final ProposedTransaction proposedTransaction) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.channelName = channelName;
        this.proposedTransaction = proposedTransaction;
    }

    @Override
    public String getTransactionId() {
        return proposedTransaction.getTransactionId();
    }

    @Override
    public byte[] getBytes() {
        return proposedTransaction.toByteArray();
    }

    @Override
    public byte[] getDigest() {
        byte[] message = proposedTransaction.getProposal().getProposalBytes().toByteArray();
        return signingIdentity.hash(message);
    }

    @Override
    public byte[] evaluate(final UnaryOperator<CallOptions> options) throws GatewayException {
        sign();
        final EvaluateRequest evaluateRequest = EvaluateRequest.newBuilder()
                .setTransactionId(proposedTransaction.getTransactionId())
                .setChannelId(channelName)
                .setProposedTransaction(proposedTransaction.getProposal())
                .addAllTargetOrganizations(proposedTransaction.getEndorsingOrganizationsList())
                .build();
        return client.evaluate(evaluateRequest, options)
                .getResult()
                .getPayload()
                .toByteArray();
    }

    @Override
    public Transaction endorse(final UnaryOperator<CallOptions> options) throws EndorseException {
        sign();
        final EndorseRequest endorseRequest = EndorseRequest.newBuilder()
                .setTransactionId(proposedTransaction.getTransactionId())
                .setChannelId(channelName)
                .setProposedTransaction(proposedTransaction.getProposal())
                .addAllEndorsingOrganizations(proposedTransaction.getEndorsingOrganizationsList())
                .build();
        EndorseResponse endorseResponse = client.endorse(endorseRequest, options);

        PreparedTransaction preparedTransaction = PreparedTransaction.newBuilder()
                .setTransactionId(proposedTransaction.getTransactionId())
                .setEnvelope(endorseResponse.getPreparedTransaction())
                .build();
        return new TransactionImpl(client, signingIdentity, preparedTransaction);
    }

    void setSignature(final byte[] signature) {
        SignedProposal signedProposal = proposedTransaction.getProposal().toBuilder()
                .setSignature(ByteString.copyFrom(signature))
                .build();

        proposedTransaction = proposedTransaction.toBuilder()
                .setProposal(signedProposal)
                .build();
    }

    private void sign() {
        if (isSigned()) {
            return;
        }

        byte[] digest = getDigest();
        byte[] signature = signingIdentity.sign(digest);
        setSignature(signature);
    }

    private boolean isSigned() {
        return !proposedTransaction.getProposal().getSignature().isEmpty();
    }
}
