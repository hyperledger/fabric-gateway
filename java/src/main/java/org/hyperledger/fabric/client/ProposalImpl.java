/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.ProposedTransaction;
import org.hyperledger.fabric.protos.peer.ProposalPackage;

final class ProposalImpl implements Proposal {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private ProposedTransaction proposedTransaction;

    ProposalImpl(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity, final ProposedTransaction proposedTransaction) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.proposedTransaction = proposedTransaction;
    }

    @Override
    public String getTransactionId() {
        return proposedTransaction.getTxId();
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
    public byte[] evaluate() {
        sign();
        return client.evaluate(proposedTransaction)
                .getValue()
                .toByteArray();
    }

    @Override
    public Transaction endorse() {
        sign();
        PreparedTransaction preparedTransaction = client.endorse(proposedTransaction);
        return new TransactionImpl(client, signingIdentity, preparedTransaction);
    }

    void setSignature(final byte[] signature) {
        ProposalPackage.SignedProposal signedProposal = proposedTransaction.getProposal().toBuilder()
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
