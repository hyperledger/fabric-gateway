/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import io.grpc.CallOptions;
import io.grpc.Deadline;
import io.grpc.StatusRuntimeException;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.protos.gateway.CommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.hyperledger.fabric.protos.peer.TxValidationCode;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.function.Function;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;

public final class SubmitTransactionTest {
    private static final TestUtils utils = TestUtils.getInstance();
    private static final Deadline defaultEndorseDeadline = Deadline.after(1, TimeUnit.DAYS);
    private static final Deadline defaultSubmitDeadline = Deadline.after(2, TimeUnit.DAYS);
    private static final Deadline defaultCommitStatusDeadline = Deadline.after(3, TimeUnit.DAYS);

    private GatewayMocker mocker;
    private GatewayServiceStub stub;
    private Gateway gateway;
    private Network network;
    private Contract contract;

    @BeforeEach
    void beforeEach() {
        mocker = new GatewayMocker();
        stub = mocker.getGatewayServiceStubSpy();

        gateway = mocker.getGatewayBuilder()
                .endorseOptions(options -> options.withDeadline(defaultEndorseDeadline))
                .submitOptions(options -> options.withDeadline(defaultSubmitDeadline))
                .commitStatusOptions(options -> options.withDeadline(defaultCommitStatusDeadline))
                .connect();
        network = gateway.getNetwork("NETWORK");
        contract = network.getContract("CHAINCODE_NAME");
    }

    @AfterEach
    void afterEach() {
        gateway.close();
        mocker.close();
    }

    @Test
    void throws_NullPointerException_on_null_transaction_name() {
        assertThatThrownBy(() -> contract.submitTransaction(null))
                .isInstanceOf(NullPointerException.class)
                .hasMessageContaining("transaction name");
    }

    @Test
    void returns_gateway_response() throws Exception {
        doReturn(utils.newEndorseResponse("MY_RESULT", "CHANNEL_NAME"))
                .when(stub).endorse(any());

        byte[] actual = contract.submitTransaction("TRANSACTION_NAME");

        assertThat(actual).asString(StandardCharsets.UTF_8).isEqualTo("MY_RESULT");
    }

    @Test
    void sends_chaincode_name() throws Exception {
        contract.submitTransaction("TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        String actual = mocker.getChaincodeSpec(request.getProposedTransaction()).getChaincodeId().getName();

        assertThat(actual).isEqualTo(contract.getChaincodeName());
    }

    @Test
    void sends_transaction_name_for_default_contract() throws Exception {
        network.getContract("CHAINCODE_NAME")
                .submitTransaction("MY_TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        List<String> chaincodeArgs = mocker.getChaincodeSpec(request.getProposedTransaction()).getInput().getArgsList().stream()
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).first().isEqualTo("MY_TRANSACTION_NAME");
    }

    @Test
    void sends_transaction_name_for_specified_contract() throws Exception {
        network.getContract("CHAINCODE_NAME", "MY_CONTRACT")
                .submitTransaction("MY_TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        List<String> chaincodeArgs = mocker.getChaincodeSpec(request.getProposedTransaction()).getInput().getArgsList().stream()
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).first().isEqualTo("MY_CONTRACT:MY_TRANSACTION_NAME");
    }

    @Test
    void sends_transaction_string_arguments() throws Exception {
        contract.submitTransaction("TRANSACTION_NAME", "one", "two", "three");

        EndorseRequest request = mocker.captureEndorse();
        List<String> chaincodeArgs = mocker.getChaincodeSpec(request.getProposedTransaction()).getInput().getArgsList().stream()
                .skip(1)
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).containsExactly("one", "two", "three");
    }

    @Test
    void sends_transaction_byte_array_arguments() throws Exception {
        byte[][] arguments = Stream.of("one", "two", "three")
                .map(s -> s.getBytes(StandardCharsets.UTF_8))
                .toArray(byte[][]::new);
        contract.submitTransaction("TRANSACTION_NAME", arguments);

        EndorseRequest request = mocker.captureEndorse();
        byte[][] chaincodeArgs = mocker.getChaincodeSpec(request.getProposedTransaction()).getInput().getArgsList().stream()
                .skip(1)
                .map(ByteString::toByteArray)
                .toArray(byte[][]::new);

        assertThat(chaincodeArgs).isDeepEqualTo(arguments);
    }

    @Test
    void sends_transient_data() throws Exception {
        contract.newProposal("TRANSACTION_NAME")
                .putTransient("uno", "one".getBytes(StandardCharsets.UTF_8))
                .putTransient("dos", "two".getBytes(StandardCharsets.UTF_8))
                .build()
                .endorse()
                .submit();

        EndorseRequest request = mocker.captureEndorse();
        Map<String, ByteString> transientData = mocker.getProposalPayload(request.getProposedTransaction()).getTransientMapMap();
        assertThat(transientData).containsOnly(
                entry("uno", ByteString.copyFrom("one", StandardCharsets.UTF_8)),
                entry("dos", ByteString.copyFrom("two", StandardCharsets.UTF_8)));
    }

    @Test
    void sets_endorsing_orgs() throws Exception {
        contract.newProposal("TRANSACTION_NAME")
                .setEndorsingOrganizations("Org1MSP", "Org3MSP")
                .build()
                .endorse()
                .submit();

        EndorseRequest request = mocker.captureEndorse();
        List<String> endorsingOrgs = request.getEndorsingOrganizationsList();
        assertThat(endorsingOrgs).containsExactlyInAnyOrder("Org1MSP", "Org3MSP");
    }

    @Test
    void uses_signer_for_endorse() throws Exception {
        Signer signer = (digest) -> "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);
        try (Gateway gateway = mocker.getGatewayBuilder().signer(signer).connect()) {
            gateway.getNetwork("NETWORK")
                    .getContract("CHAINCODE_NAME")
                    .submitTransaction("TRANSACTION_NAME");

            EndorseRequest request = mocker.captureEndorse();
            String signature = request.getProposedTransaction().getSignature().toStringUtf8();

            assertThat(signature).isEqualTo("MY_SIGNATURE");
        }
    }

    @Test
    void uses_identity() throws Exception {
        Identity identity = new X509Identity("MY_MSP_ID", utils.getCredentials().getCertificate());
        try (Gateway gateway = mocker.getGatewayBuilder().identity(identity).connect()) {
            gateway.getNetwork("NETWORK")
                    .getContract("CHAINCODE_NAME")
                    .submitTransaction("TRANSACTION_NAME");

            EndorseRequest request = mocker.captureEndorse();
            ByteString serializedIdentity = mocker.getSignatureHeader(request.getProposedTransaction()).getCreator();

            byte[] expected = GatewayUtils.serializeIdentity(identity);
            assertThat(serializedIdentity.toByteArray()).isEqualTo(expected);
        }
    }

    @Test
    void sends_network_name_in_proposal_for_default_contract() throws Exception {
       network.getContract("CHAINCODE_NAME")
                .submitTransaction("TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        String networkName = mocker.getChannelHeader(request.getProposedTransaction()).getChannelId();

        assertThat(networkName).isEqualTo(network.getName());
    }

    @Test
    void sends_network_name_in_proposed_transaction_for_default_contract() throws Exception {
        network.getContract("CHAINCODE_NAME")
                .submitTransaction("TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        String networkName = request.getChannelId();

        assertThat(networkName).isEqualTo(network.getName());
    }

    @Test
    void sends_network_name_in_proposal_for_specified_contract() throws Exception {
        network.getContract("CHAINCODE_NAME", "CONTRACT_NAME")
                .submitTransaction("TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        String networkName = mocker.getChannelHeader(request.getProposedTransaction()).getChannelId();

        assertThat(networkName).isEqualTo(network.getName());
    }

    @Test
    void sends_network_name_in_proposed_transaction_for_specified_contract() throws Exception {
        network.getContract("CHAINCODE_NAME", "CONTRACT_NAME")
                .submitTransaction("TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        String networkName = request.getChannelId();

        assertThat(networkName).isEqualTo(network.getName());
    }

    @Test
    void sends_transaction_ID_in_proposed_transaction() throws Exception {
        Proposal proposal = contract.newProposal("TRANSACTION_NAME").build();
        proposal.endorse().submit();

        String expected = proposal.getTransactionId();
        assertThat(expected).isNotEmpty();

        EndorseRequest request = mocker.captureEndorse();
        String actual = request.getTransactionId();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void proposals_built_by_same_builder_have_different_transaction_IDs() {
        Proposal.Builder builder = contract.newProposal("TRANSACTION_NAME");

        Proposal proposal1 = builder.build();
        Proposal proposal2 = builder.build();

        assertThat(proposal1.getTransactionId()).isNotEqualTo(proposal2.getTransactionId());
    }

    @Test
    void uses_signer_for_submit() throws Exception {
        Signer signer = (digest) -> "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);
        try (Gateway gateway = mocker.getGatewayBuilder().signer(signer).connect()) {
            gateway.getNetwork("NETWORK")
                    .getContract("CHAINCODE_NAME")
                    .submitTransaction("TRANSACTION_NAME");

            SubmitRequest request = mocker.captureSubmit();
            String signature = request.getPreparedTransaction().getSignature().toStringUtf8();

            assertThat(signature).isEqualTo("MY_SIGNATURE");
        }
    }

    @Test
    void uses_hash() throws Exception {
        List<String> actual = new ArrayList<>();
        Function<byte[], byte[]> hash = message -> "MY_DIGEST".getBytes(StandardCharsets.UTF_8);
        Signer signer = digest -> {
            actual.add(new String(digest, StandardCharsets.UTF_8));
            return "SIGNATURE".getBytes(StandardCharsets.UTF_8);
        };

        try (Gateway gateway = mocker.getGatewayBuilder().hash(hash).signer(signer).connect()) {
            gateway.getNetwork("NETWORK")
                    .getContract("CHAINCODE_NAME")
                    .submitTransaction("TRANSACTION_NAME");

            assertThat(actual).hasSameElementsAs(Arrays.asList("MY_DIGEST", "MY_DIGEST"));
        }
    }

    @Test
    void throws_on_endorse_connection_error() {
        StatusRuntimeException expected = new StatusRuntimeException(io.grpc.Status.UNAVAILABLE);
        doThrow(expected).when(stub).endorse(any());

        Proposal proposal = contract.newProposal("TRANSACTION_NAME").build();

        EndorseException e = catchThrowableOfType(proposal::endorse, EndorseException.class);
        assertThat(e.getTransactionId()).isEqualTo(proposal.getTransactionId());
        assertThat(e.getStatus()).isEqualTo(expected.getStatus());
        assertThat(e).hasCauseInstanceOf(StatusRuntimeException.class);
    }

    @Test
    void throws_on_submit_connection_error() throws EndorseException {
        StatusRuntimeException expected = new StatusRuntimeException(io.grpc.Status.UNAVAILABLE);
        doThrow(expected).when(stub).submit(any());

        Transaction transaction = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse();

        SubmitException e = catchThrowableOfType(transaction::submit, SubmitException.class);
        assertThat(e.getTransactionId()).isEqualTo(transaction.getTransactionId());
        assertThat(e.getStatus()).isEqualTo(expected.getStatus());
        assertThat(e).hasCauseInstanceOf(StatusRuntimeException.class);
    }

    @Test
    void throws_on_commit_status_connection_error() throws EndorseException, SubmitException {
        StatusRuntimeException expected = new StatusRuntimeException(io.grpc.Status.UNAVAILABLE);
        doThrow(expected).when(stub).commitStatus(any());

        SubmittedTransaction commit = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submitAsync();

        CommitStatusException e = catchThrowableOfType(commit::getStatus, CommitStatusException.class);
        assertThat(e.getTransactionId()).isEqualTo(commit.getTransactionId());
        assertThat(e.getStatus()).isEqualTo(expected.getStatus());
        assertThat(e).hasCauseInstanceOf(StatusRuntimeException.class);
    }

    @Test
    void throws_on_commit_failure() throws EndorseException {
        doReturn(utils.newCommitStatusResponse(TxValidationCode.MVCC_READ_CONFLICT))
                .when(stub).commitStatus(any());

        Transaction transaction = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse();

        CommitException e = catchThrowableOfType(transaction::submit, CommitException.class);

        assertThat(e).hasMessageContaining(TxValidationCode.MVCC_READ_CONFLICT.name());
        assertThat(e.getCode()).isEqualTo(TxValidationCode.MVCC_READ_CONFLICT);
        assertThat(e.getTransactionId()).isEqualTo(transaction.getTransactionId());
    }

    @Test
    void sends_transaction_ID_in_commit_status_request() throws Exception {
        Proposal proposal = contract.newProposal("TRANSACTION_NAME").build();
        proposal.endorse().submit();

        String expected = proposal.getTransactionId();
        assertThat(expected).isNotEmpty();

        SignedCommitStatusRequest signedRequest = mocker.captureCommitStatus();
        CommitStatusRequest request = CommitStatusRequest.parseFrom(signedRequest.getRequest());
        String actual = request.getTransactionId();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void sends_network_name_in_commit_status_request() throws Exception {
        contract.submitTransaction("TRANSACTION_NAME");

        SignedCommitStatusRequest signedRequest = mocker.captureCommitStatus();
        CommitStatusRequest request = CommitStatusRequest.parseFrom(signedRequest.getRequest());
        String networkName = request.getChannelId();

        assertThat(networkName).isEqualTo(network.getName());
    }

    @Test
    void commit_returns_transaction_validation_code() throws EndorseException, SubmitException, CommitStatusException {
        doReturn(utils.newCommitStatusResponse(TxValidationCode.MVCC_READ_CONFLICT))
                .when(stub).commitStatus(any());

        Status status = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submitAsync()
                .getStatus();

        assertThat(status.getCode()).isEqualTo(TxValidationCode.MVCC_READ_CONFLICT);
    }

    @Test
    void commit_returns_successful_for_successful_transaction() throws EndorseException, SubmitException, CommitStatusException {
        Status status = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submitAsync()
                .getStatus();

        assertThat(status.isSuccessful()).isTrue();
    }

    @Test
    void commit_returns_unsuccessful_for_failed_transaction() throws EndorseException, SubmitException, CommitStatusException {
        doReturn(utils.newCommitStatusResponse(TxValidationCode.MVCC_READ_CONFLICT))
                .when(stub).commitStatus(any());

        Status status = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submitAsync()
                .getStatus();

        assertThat(status.isSuccessful()).isFalse();
    }

    @Test
    void commit_returns_block_number() throws EndorseException, SubmitException, CommitStatusException {
        doReturn(utils.newCommitStatusResponse(TxValidationCode.MVCC_READ_CONFLICT, 101))
                .when(stub).commitStatus(any());

        Status status = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submitAsync()
                .getStatus();

        assertThat(status.getBlockNumber()).isEqualTo(101);
    }

    @Test
    @SuppressWarnings("deprecation")
    void endorse_uses_legacy_specified_call_options() throws EndorseException {
        Deadline expected = Deadline.after(1, TimeUnit.MINUTES);

        contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse(CallOption.deadline(expected));

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual)
                .first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(expected);
    }

    @Test
    void endorse_uses_specified_call_options() throws EndorseException {
        Deadline expected = Deadline.after(1, TimeUnit.MINUTES);

        contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse(options -> options.withDeadline(expected));

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual)
                .first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(expected);
    }

    @Test
    @SuppressWarnings("deprecation")
    void endorse_uses_legacy_default_call_options() throws EndorseException {
        Deadline expected = Deadline.after(1, TimeUnit.MINUTES);

        try (Gateway gateway = mocker.getGatewayBuilder()
                .endorseOptions(CallOption.deadline(expected))
                .connect()) {
            gateway.getNetwork("NETWORK")
                    .getContract("CHAINCODE_NAME")
                    .newProposal("TRANSACTION_NAME")
                    .build()
                    .endorse();
        }

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual)
                .first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(expected);
    }

    @Test
    void endorse_uses_default_call_options() throws EndorseException {
        contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse();

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual)
                .first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(defaultEndorseDeadline);
    }

    @Test
    @SuppressWarnings("deprecation")
    void submit_uses_legacy_specified_call_options_for_submit_and_commitStatus()
            throws CommitException, EndorseException, SubmitException, CommitStatusException {
        Deadline expected = Deadline.after(1, TimeUnit.MINUTES);
        Transaction transaction = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse();
        mocker.reset();

        transaction.submit(CallOption.deadline(expected));

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual)
                .hasSize(2)
                .extracting(CallOptions::getDeadline)
                .containsOnly(expected);
    }

    @Test
    void submit_uses_specified_call_options_for_submit_and_commitStatus() throws CommitException, EndorseException, SubmitException, CommitStatusException {
        Deadline expected = Deadline.after(1, TimeUnit.MINUTES);
        Transaction transaction = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse();
        mocker.reset();

        transaction.submit(options -> options.withDeadline(expected));

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual)
                .hasSize(2)
                .extracting(CallOptions::getDeadline)
                .containsOnly(expected);
    }

    @Test
    @SuppressWarnings("deprecation")
    void submit_uses_legacy_default_call_options_for_submit_and_commitStatus() throws CommitException, EndorseException, SubmitException,
            CommitStatusException {
        Deadline submitDeadline = Deadline.after(1, TimeUnit.MINUTES);
        Deadline commitStatusDeadline = Deadline.after(2, TimeUnit.MINUTES);

        try (Gateway gateway = mocker.getGatewayBuilder()
                .submitOptions(CallOption.deadline(submitDeadline))
                .commitStatusOptions(CallOption.deadline(commitStatusDeadline))
                .connect()) {
            Transaction transaction = gateway.getNetwork("NETWORK")
                    .getContract("CHAINCODE_NAME").newProposal("TRANSACTION_NAME")
                    .build()
                    .endorse();
            mocker.reset();

            transaction.submit();
        }

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual)
                .hasSize(2)
                .extracting(CallOptions::getDeadline)
                .containsExactly(submitDeadline, commitStatusDeadline);
    }

    @Test
    void submit_uses_default_call_options_for_submit_and_commitStatus() throws CommitException, EndorseException, SubmitException, CommitStatusException {
        Transaction transaction = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse();
        mocker.reset();

        transaction.submit();

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual)
                .hasSize(2)
                .extracting(CallOptions::getDeadline)
                .containsExactly(defaultSubmitDeadline, defaultCommitStatusDeadline);
    }
}
