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
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.EvaluateRequest;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

public final class OfflineSignTest {
    private static final TestUtils utils = TestUtils.getInstance();

    private GatewayMocker mocker;
    private Gateway gateway;
    private Network network;
    private Contract contract;

    @BeforeEach
    void beforeEach() throws Exception {
        mocker = new GatewayMocker(newBuilderWithoutSigner());
        gateway = mocker.getGatewayBuilder().connect();
        network = gateway.getNetwork("NETWORK");
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
    void newProposal_throws_NullPointerException_on_null_transaction_name() {
        assertThatThrownBy(() -> contract.newProposal(null))
                .isInstanceOf(NullPointerException.class)
                .hasMessageContaining("transaction name");
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

        EvaluateRequest request = mocker.captureEvaluate();
        byte[] actual = request.getProposedTransaction().getSignature().toByteArray();

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

        EndorseRequest request = mocker.captureEndorse();
        byte[] actual = request.getProposedTransaction().getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void submit_throws_with_no_signer_and_no_explicit_signing() throws InvalidProtocolBufferException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction transaction = signedProposal.endorse();

        assertThatThrownBy(() -> transaction.submitAsync())
                .isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void submit_uses_offline_signature() throws InvalidProtocolBufferException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), expected);
        signedTransaction.submitAsync();

        SubmitRequest request = mocker.captureSubmit();
        byte[] actual = request.getPreparedTransaction().getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void commit_throws_with_no_signer_and_no_explicit_signing() throws InvalidProtocolBufferException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction transaction = signedProposal.endorse();
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Commit commit = signedTransaction.submitAsync();

        assertThatThrownBy(() -> commit.getStatus())
                .isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void commit_uses_offline_signature() throws InvalidProtocolBufferException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction transaction = signedProposal.endorse();
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Commit unsignedCommit = signedTransaction.submitAsync();
        Commit signedCommit = network.newSignedCommit(unsignedCommit.getBytes(), expected);
        signedCommit.getStatus();

        SignedCommitStatusRequest request = mocker.captureCommitStatus();
        byte[] actual = request.getSignature().toByteArray();

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
    void signed_transaction_keeps_same_transaction_ID() throws InvalidProtocolBufferException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        String expected = unsignedTransaction.getTransactionId();

        Transaction signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        String actual = signedTransaction.getTransactionId();

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

    @Test
    void signed_commit_keeps_same_transaction_ID() throws InvalidProtocolBufferException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Commit unsignedCommit = signedTransaction.submitAsync();
        String expected = unsignedCommit.getTransactionId();

        Commit signedCommit = network.newSignedCommit(unsignedCommit.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        String actual = signedCommit.getTransactionId();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_commit_keeps_same_digest() throws InvalidProtocolBufferException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = contract.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Commit unsignedCommit = signedTransaction.submitAsync();
        byte[] expected = unsignedCommit.getDigest();

        Commit signedCommit = network.newSignedCommit(unsignedCommit.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedCommit.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void chaincode_events_uses_offline_signature() throws Exception {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        ChaincodeEventsRequest unsignedRequest = network.newChaincodeEventsRequest("CHAINCODE_ID").build();
        ChaincodeEventsRequest signedRequest = network.newSignedChaincodeEventsRequest(unsignedRequest.getBytes(), expected);
        try (CloseableIterator<ChaincodeEvent> iter = signedRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> {});
        }

        SignedChaincodeEventsRequest request = mocker.captureChaincodeEvents();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_chaincode_events_keep_same_digest() throws Exception {
        ChaincodeEventsRequest unsignedRequest = network.newChaincodeEventsRequest("CHAINCODE_ID").build();
        byte[] expected = unsignedRequest.getDigest();

        ChaincodeEventsRequest signedRequest = network.newSignedChaincodeEventsRequest(unsignedRequest.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedRequest.getDigest();

        assertThat(actual).isEqualTo(expected);
    }
}
