/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.nio.charset.StandardCharsets;

import com.google.protobuf.InvalidProtocolBufferException;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.ProposedTransaction;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

public final class OfflineSignTest {
    private static final TestUtils utils = TestUtils.getInstance();

    private GatewayMocker mocker;
    private Gateway gateway;
    private Contract contract;

    @BeforeEach
    void beforeEach() {
        mocker = new GatewayMocker(newBuilderWithoutSigner());
        gateway = mocker.getGatewayBuilder().connect();
        Network network = gateway.getNetwork("NETWORK");
        contract = network.getContract("CHAINCODE_ID");
    }

    private Gateway.Builder newBuilderWithoutSigner() {
        Gateway.Builder builder = Gateway.newInstance();
        Identity identity = new X509Identity("MSP_ID", utils.getCredentials().getCertificate());
        builder.identity(identity);
        return builder;
    }

    @AfterEach
    void afterEach() {
        gateway.close();
        mocker.close();
    }

    @Test
    void evaluate_throws_with_no_signer_and_no_explicit_signing() {
        Proposal proposal = contract.newProposal("TRANSACTION_NAME").build();

        assertThatThrownBy(proposal::evaluate)
                .isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void evaluate_uses_offline_signature() throws InvalidProtocolBufferException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), expected);
        signedProposal.evaluate();

        ProposedTransaction request = mocker.captureEvaluate();
        byte[] actual = request.getProposal().getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void endorse_throws_with_no_signer_and_no_explicit_signing() {
        Proposal proposal = contract.newProposal("TRANSACTION_NAME").build();

        assertThatThrownBy(proposal::endorse)
                .isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void endorse_uses_offline_signature() throws InvalidProtocolBufferException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), expected);
        signedProposal.endorse();

        ProposedTransaction request = mocker.captureEndorse();
        byte[] actual = request.getProposal().getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void submit_throws_with_no_signer_and_no_explicit_signing() throws InvalidProtocolBufferException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction transaction = signedProposal.endorse();

        assertThatThrownBy(transaction::submitSync)
                .isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void submit_uses_offline_signature() throws InvalidProtocolBufferException, ContractException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), expected);
        signedTransaction.submitSync();

        PreparedTransaction request = mocker.captureSubmit();
        byte[] actual = request.getEnvelope().getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_proposal_keeps_same_transaction_ID() throws InvalidProtocolBufferException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        String expected = unsignedProposal.getTransactionId();

        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        String actual = signedProposal.getTransactionId();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_proposal_keeps_same_digest() throws InvalidProtocolBufferException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        byte[] expected = unsignedProposal.getDigest();

        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedProposal.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_transaction_keeps_same_digest() throws InvalidProtocolBufferException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        byte[] expected = unsignedTransaction.getDigest();

        Transaction signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedTransaction.getDigest();

        assertThat(actual).isEqualTo(expected);
    }
}
