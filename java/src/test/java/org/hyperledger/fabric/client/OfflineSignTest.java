/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.nio.charset.StandardCharsets;
import java.util.List;

import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.protos.common.Envelope;
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
    void beforeEach() {
        mocker = new GatewayMocker(newBuilderWithoutSigner());
        gateway = mocker.getGatewayBuilder().connect();
        network = gateway.getNetwork("NETWORK");
        contract = network.getContract("CHAINCODE_NAME");
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
    void evaluate_uses_offline_signature() throws GatewayException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), expected);
        signedProposal.evaluate();

        EvaluateRequest request = mocker.captureEvaluate();
        byte[] actual = request.getProposedTransaction().getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void evaluate_retains_signature() throws GatewayException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), expected);
        Proposal newProposal = gateway.newProposal(signedProposal.getBytes());

        newProposal.evaluate();

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
    void endorse_uses_offline_signature() throws EndorseException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), expected);
        signedProposal.endorse();

        EndorseRequest request = mocker.captureEndorse();
        byte[] actual = request.getProposedTransaction().getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void endorse_retains_signature() throws EndorseException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), expected);
        Proposal newProposal = gateway.newProposal(signedProposal.getBytes());
        newProposal.endorse();

        EndorseRequest request = mocker.captureEndorse();
        byte[] actual = request.getProposedTransaction().getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void submit_throws_with_no_signer_and_no_explicit_signing() throws EndorseException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction transaction = signedProposal.endorse();

        assertThatThrownBy(transaction::submitAsync)
                .isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void submit_uses_offline_signature() throws EndorseException, SubmitException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), expected);
        signedTransaction.submitAsync();

        SubmitRequest request = mocker.captureSubmit();
        byte[] actual = request.getPreparedTransaction().getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void submit_retains_signature() throws EndorseException, SubmitException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), expected);
        Transaction newTransaction = gateway.newTransaction(signedTransaction.getBytes());
        newTransaction.submitAsync();

        SubmitRequest request = mocker.captureSubmit();
        byte[] actual = request.getPreparedTransaction().getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void commit_throws_with_no_signer_and_no_explicit_signing() throws EndorseException, SubmitException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Commit commit = signedTransaction.submitAsync();

        assertThatThrownBy(commit::getStatus)
                .isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void commit_uses_offline_signature() throws EndorseException, SubmitException, CommitStatusException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Commit unsignedCommit = signedTransaction.submitAsync();
        Commit signedCommit = gateway.newSignedCommit(unsignedCommit.getBytes(), expected);
        signedCommit.getStatus();

        SignedCommitStatusRequest request = mocker.captureCommitStatus();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void commit_retains_signature() throws EndorseException, SubmitException, CommitStatusException {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Commit unsignedCommit = signedTransaction.submitAsync();
        Commit signedCommit = gateway.newSignedCommit(unsignedCommit.getBytes(), expected);
        Commit newCommit = gateway.newCommit(signedCommit.getBytes());
        newCommit.getStatus();

        SignedCommitStatusRequest request = mocker.captureCommitStatus();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_proposal_keeps_same_transaction_ID() {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        String expected = unsignedProposal.getTransactionId();

        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        String actual = signedProposal.getTransactionId();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_proposal_keeps_same_digest() {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        byte[] expected = unsignedProposal.getDigest();

        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedProposal.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void proposal_keeps_same_digest_during_deserialization() {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        byte[] expected = unsignedProposal.getDigest();

        Proposal newProposal = gateway.newProposal(unsignedProposal.getBytes());
        byte[] actual = newProposal.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_proposal_keeps_same_endorsing_orgs() throws GatewayException {
        Contract contract = network.getContract("CHAINCODE_NAME");
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME")
                .setEndorsingOrganizations("Org1MSP", "Org3MSP")
                .build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        signedProposal.evaluate();

        EvaluateRequest request = mocker.captureEvaluate();
        List<String> endorsingOrgs = request.getTargetOrganizationsList();
        assertThat(endorsingOrgs).containsExactlyInAnyOrder("Org1MSP", "Org3MSP");
    }



    @Test
    void signed_transaction_keeps_same_transaction_ID() throws EndorseException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        String expected = unsignedTransaction.getTransactionId();

        Transaction signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        String actual = signedTransaction.getTransactionId();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_transaction_keeps_same_digest() throws EndorseException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        byte[] expected = unsignedTransaction.getDigest();

        Transaction signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedTransaction.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void transaction_keeps_same_digest_during_deserialization() throws EndorseException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        byte[] expected = unsignedTransaction.getDigest();

        Transaction newTransaction = gateway.newTransaction(unsignedTransaction.getBytes());
        byte[] actual = newTransaction.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_commit_keeps_same_transaction_ID() throws EndorseException, SubmitException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Commit unsignedCommit = signedTransaction.submitAsync();
        String expected = unsignedCommit.getTransactionId();

        Commit signedCommit = gateway.newSignedCommit(unsignedCommit.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        String actual = signedCommit.getTransactionId();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_commit_keeps_same_digest() throws EndorseException, SubmitException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Commit unsignedCommit = signedTransaction.submitAsync();
        byte[] expected = unsignedCommit.getDigest();

        Commit signedCommit = gateway.newSignedCommit(unsignedCommit.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedCommit.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void commit_keeps_same_digest_during_deserialization() throws EndorseException, SubmitException {
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME").build();
        Proposal signedProposal = gateway.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = gateway.newSignedTransaction(unsignedTransaction.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        Commit unsignedCommit = signedTransaction.submitAsync();
        byte[] expected = unsignedCommit.getDigest();

        Commit newCommit = gateway.newCommit(unsignedCommit.getBytes());
        byte[] actual = newCommit.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void chaincode_events_throws_with_no_signer_and_no_explicit_signing() {
        ChaincodeEventsRequest unsignedRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME").build();
        assertThatThrownBy(unsignedRequest::getEvents).isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void chaincode_events_uses_offline_signature() {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        ChaincodeEventsRequest unsignedRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME").build();
        ChaincodeEventsRequest signedRequest = gateway.newSignedChaincodeEventsRequest(unsignedRequest.getBytes(), expected);
        try (CloseableIterator<?> iter = signedRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> { });
        }

        SignedChaincodeEventsRequest request = mocker.captureChaincodeEvents();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void chaincode_events_retains_signature() {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        ChaincodeEventsRequest unsignedRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME").build();
        ChaincodeEventsRequest signedRequest = gateway.newSignedChaincodeEventsRequest(unsignedRequest.getBytes(), expected);
        ChaincodeEventsRequest newChaincodeRequest = gateway.newChaincodeEventsRequest(signedRequest.getBytes());
        try (CloseableIterator<?> iter = newChaincodeRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> { });
        }

        SignedChaincodeEventsRequest request = mocker.captureChaincodeEvents();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_chaincode_events_keeps_same_digest() {
        ChaincodeEventsRequest unsignedRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME").build();
        byte[] expected = unsignedRequest.getDigest();

        ChaincodeEventsRequest signedRequest = gateway.newSignedChaincodeEventsRequest(unsignedRequest.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedRequest.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void chaincode_events_keeps_same_digest_during_deserialization() {
        ChaincodeEventsRequest unsignedRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME").build();
        byte[] expected = unsignedRequest.getDigest();

        ChaincodeEventsRequest newChaincodeEventsRequest = gateway.newChaincodeEventsRequest(unsignedRequest.getBytes());
        byte[] actual = newChaincodeEventsRequest.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void block_events_throws_with_no_signer_and_no_explicit_signing() {
        BlockEventsRequest unsignedRequest = network.newBlockEventsRequest().build();
        assertThatThrownBy(unsignedRequest::getEvents).isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void block_events_uses_offline_signature() {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        BlockEventsRequest unsignedRequest = network.newBlockEventsRequest().build();
        BlockEventsRequest signedRequest = gateway.newSignedBlockEventsRequest(unsignedRequest.getBytes(), expected);
        try (CloseableIterator<?> iter = signedRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> { });
        }

        Envelope request = mocker.captureBlockEvents().findFirst().get();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void block_events_retains_signature() {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        BlockEventsRequest unsignedRequest = network.newBlockEventsRequest().build();
        BlockEventsRequest signedRequest = gateway.newSignedBlockEventsRequest(unsignedRequest.getBytes(), expected);
        BlockEventsRequest newBlockEventRequest = gateway.newBlockEventsRequest(signedRequest.getBytes());
        try (CloseableIterator<?> iter = newBlockEventRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> { });
        }

        Envelope request = mocker.captureBlockEvents().findFirst().get();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_block_events_keeps_same_digest() {
        BlockEventsRequest unsignedRequest = network.newBlockEventsRequest().build();
        byte[] expected = unsignedRequest.getDigest();

        BlockEventsRequest signedRequest = gateway.newSignedBlockEventsRequest(unsignedRequest.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedRequest.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void block_events_keeps_same_digest_during_deserialization() {
        BlockEventsRequest unsignedRequest = network.newBlockEventsRequest().build();
        byte[] expected = unsignedRequest.getDigest();

        BlockEventsRequest newBlockEventsRequest = gateway.newBlockEventsRequest(unsignedRequest.getBytes());
        byte[] actual = newBlockEventsRequest.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void filtered_block_events_throws_with_no_signer_and_no_explicit_signing() {
        FilteredBlockEventsRequest unsignedRequest = network.newFilteredBlockEventsRequest().build();
        assertThatThrownBy(unsignedRequest::getEvents).isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void filtered_block_events_uses_offline_signature() {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        FilteredBlockEventsRequest unsignedRequest = network.newFilteredBlockEventsRequest().build();
        FilteredBlockEventsRequest signedRequest = gateway.newSignedFilteredBlockEventsRequest(unsignedRequest.getBytes(), expected);
        try (CloseableIterator<?> iter = signedRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> { });
        }

        Envelope request = mocker.captureFilteredBlockEvents().findFirst().get();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void filtered_block_events_retains_signature() {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        FilteredBlockEventsRequest unsignedRequest = network.newFilteredBlockEventsRequest().build();
        FilteredBlockEventsRequest signedRequest = gateway.newSignedFilteredBlockEventsRequest(unsignedRequest.getBytes(), expected);
        FilteredBlockEventsRequest newFilteredBlockEventsRequest = gateway.newFilteredBlockEventsRequest(signedRequest.getBytes());

        try (CloseableIterator<?> iter = newFilteredBlockEventsRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> { });
        }

        Envelope request = mocker.captureFilteredBlockEvents().findFirst().get();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_filtered_block_events_keeps_same_digest() {
        FilteredBlockEventsRequest unsignedRequest = network.newFilteredBlockEventsRequest().build();
        byte[] expected = unsignedRequest.getDigest();

        FilteredBlockEventsRequest signedRequest = gateway.newSignedFilteredBlockEventsRequest(unsignedRequest.getBytes(),
                "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedRequest.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void filtered_block_events_keeps_same_digest_during_deserialization() {
        FilteredBlockEventsRequest unsignedRequest = network.newFilteredBlockEventsRequest().build();
        byte[] expected = unsignedRequest.getDigest();

        FilteredBlockEventsRequest newRequest = gateway.newFilteredBlockEventsRequest(unsignedRequest.getBytes());
        byte[] actual = newRequest.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void block_and_private_data_events_throws_with_no_signer_and_no_explicit_signing() {
        BlockAndPrivateDataEventsRequest unsignedRequest = network.newBlockAndPrivateDataEventsRequest().build();
        assertThatThrownBy(unsignedRequest::getEvents).isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void block_and_private_data_events_uses_offline_signature() {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        BlockAndPrivateDataEventsRequest unsignedRequest = network.newBlockAndPrivateDataEventsRequest().build();
        BlockAndPrivateDataEventsRequest signedRequest = gateway.newSignedBlockAndPrivateDataEventsRequest(unsignedRequest.getBytes(), expected);
        try (CloseableIterator<?> iter = signedRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> { });
        }

        Envelope request = mocker.captureBlockAndPrivateDataEvents().findFirst().get();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void block_and_private_data_events_retains_signature() {
        byte[] expected = "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);

        BlockAndPrivateDataEventsRequest unsignedRequest = network.newBlockAndPrivateDataEventsRequest().build();
        BlockAndPrivateDataEventsRequest signedRequest = gateway.newSignedBlockAndPrivateDataEventsRequest(unsignedRequest.getBytes(), expected);
        BlockAndPrivateDataEventsRequest newBlockAndPrivateDataEventsRequest = gateway.newBlockAndPrivateDataEventsRequest(signedRequest.getBytes());

        try (CloseableIterator<?> iter = newBlockAndPrivateDataEventsRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> { });
        }

        Envelope request = mocker.captureBlockAndPrivateDataEvents().findFirst().get();
        byte[] actual = request.getSignature().toByteArray();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void signed_block_and_private_data_events_keeps_same_digest() {
        BlockAndPrivateDataEventsRequest unsignedRequest = network.newBlockAndPrivateDataEventsRequest().build();
        byte[] expected = unsignedRequest.getDigest();

        BlockAndPrivateDataEventsRequest signedRequest = gateway.newSignedBlockAndPrivateDataEventsRequest(unsignedRequest.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        byte[] actual = signedRequest.getDigest();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void block_and_private_data_events_keeps_same_digest_during_deserialization() {
        BlockAndPrivateDataEventsRequest unsignedRequest = network.newBlockAndPrivateDataEventsRequest().build();
        byte[] expected = unsignedRequest.getDigest();

        BlockAndPrivateDataEventsRequest newBlockAndPrivateDataEventsRequest = gateway.newBlockAndPrivateDataEventsRequest(unsignedRequest.getBytes());
        byte[] actual = newBlockAndPrivateDataEventsRequest.getDigest();

        assertThat(actual).isEqualTo(expected);
    }
}
