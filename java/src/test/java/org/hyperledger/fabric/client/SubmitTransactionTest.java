/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.function.Function;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import com.google.protobuf.ByteString;
import io.grpc.StatusRuntimeException;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.protos.gateway.CommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.assertj.core.api.Assertions.catchThrowableOfType;
import static org.assertj.core.api.Assertions.entry;
import static org.hyperledger.fabric.protos.peer.TransactionPackage.TxValidationCode;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;

public final class SubmitTransactionTest {
    private static final TestUtils utils = TestUtils.getInstance();

    private GatewayMocker mocker;
    private GatewayServiceStub stub;
    private Gateway gateway;
    private Network network;

    @BeforeEach
    void beforeEach() {
        mocker = new GatewayMocker();
        stub = mocker.getServiceStubSpy();

        gateway = mocker.getGatewayBuilder().connect();
        network = gateway.getNetwork("NETWORK");
    }

    @AfterEach
    void afterEach() {
        gateway.close();
        mocker.close();
    }

    @Test
    void throws_NullPointerException_on_null_transaction_name() {
        Contract contract = network.getContract("CHAINCODE_NAME", "CONTRACT_NAME");

        assertThatThrownBy(() -> contract.submitTransaction(null))
                .isInstanceOf(NullPointerException.class)
                .hasMessageContaining("transaction name");
    }

    @Test
    void returns_gateway_response() throws Exception {
        doReturn(utils.newEndorseResponse("MY_RESULT"))
                .when(stub).endorse(any());

        Contract contract = network.getContract("CHAINCODE_NAME");
        byte[] actual = contract.submitTransaction("TRANSACTION_NAME");

        assertThat(actual).asString(StandardCharsets.UTF_8).isEqualTo("MY_RESULT");
    }

    @Test
    void sends_chaincode_name() throws Exception {
        Contract contract = network.getContract("MY_CHAINCODE_NAME");
        contract.submitTransaction("TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        String actual = mocker.getChaincodeSpec(request.getProposedTransaction()).getChaincodeId().getName();

        assertThat(actual).isEqualTo("MY_CHAINCODE_NAME");
    }

    @Test
    void sends_transaction_name_for_default_contract() throws Exception {
        Contract contract = network.getContract("CHAINCODE_NAME");
        contract.submitTransaction("MY_TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        List<String> chaincodeArgs = mocker.getChaincodeSpec(request.getProposedTransaction()).getInput().getArgsList().stream()
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).first().isEqualTo("MY_TRANSACTION_NAME");
    }

    @Test
    void sends_transaction_name_for_specified_contract() throws Exception {
        Contract contract = network.getContract("CHAINCODE_NAME", "MY_CONTRACT");
        contract.submitTransaction("MY_TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        List<String> chaincodeArgs = mocker.getChaincodeSpec(request.getProposedTransaction()).getInput().getArgsList().stream()
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).first().isEqualTo("MY_CONTRACT:MY_TRANSACTION_NAME");
    }

    @Test
    void sends_transaction_string_arguments() throws Exception {
        Contract contract = network.getContract("CHAINCODE_NAME");
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
        Contract contract = network.getContract("CHAINCODE_NAME");
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
        Contract contract = network.getContract("CHAINCODE_NAME");
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
        Contract contract = network.getContract("CHAINCODE_NAME");
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
    void sets_endorsing_orgs_offline_signing() throws Exception {
        Contract contract = network.getContract("CHAINCODE_NAME");
        Proposal unsignedProposal = contract.newProposal("TRANSACTION_NAME")
                .setEndorsingOrganizations("Org1MSP", "Org3MSP")
                .build();
        Proposal signedProposal = contract.newSignedProposal(unsignedProposal.getBytes(), "SIGNATURE".getBytes(StandardCharsets.UTF_8));
        signedProposal.endorse().submit();

        EndorseRequest request = mocker.captureEndorse();
        List<String> endorsingOrgs = request.getEndorsingOrganizationsList();
        assertThat(endorsingOrgs).containsExactlyInAnyOrder("Org1MSP", "Org3MSP");
    }

    @Test
    void uses_signer_for_endorse() throws Exception {
        Signer signer = (digest) -> "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);
        try (Gateway gateway = mocker.getGatewayBuilder().signer(signer).connect()) {
            network = gateway.getNetwork("NETWORK");

            Contract contract = network.getContract("CHAINCODE_NAME");
            contract.submitTransaction("TRANSACTION_NAME");

            EndorseRequest request = mocker.captureEndorse();
            String signature = request.getProposedTransaction().getSignature().toStringUtf8();

            assertThat(signature).isEqualTo("MY_SIGNATURE");
        }
    }

    @Test
    void uses_identity() throws Exception {
        Identity identity = new X509Identity("MY_MSP_ID", utils.getCredentials().getCertificate());
        try (Gateway gateway = mocker.getGatewayBuilder().identity(identity).connect()) {
            network = gateway.getNetwork("NETWORK");

            Contract contract = network.getContract("CHAINCODE_NAME");
            contract.submitTransaction("TRANSACTION_NAME");

            EndorseRequest request = mocker.captureEndorse();
            ByteString serializedIdentity = mocker.getSignatureHeader(request.getProposedTransaction()).getCreator();

            byte[] expected = GatewayUtils.serializeIdentity(identity);
            assertThat(serializedIdentity.toByteArray()).isEqualTo(expected);
        }
    }

    @Test
    void sends_network_name_in_proposal() throws Exception {
        network = gateway.getNetwork("MY_NETWORK");

        Contract contract = network.getContract("CHAINCODE_NAME");
        contract.submitTransaction("TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        String networkName = mocker.getChannelHeader(request.getProposedTransaction()).getChannelId();

        assertThat(networkName).isEqualTo("MY_NETWORK");
    }

    @Test
    void sends_network_name_in_proposed_transaction() throws Exception {
        network = gateway.getNetwork("MY_NETWORK");

        Contract contract = network.getContract("CHAINCODE_NAME");
        contract.submitTransaction("TRANSACTION_NAME");

        EndorseRequest request = mocker.captureEndorse();
        String networkName = request.getChannelId();

        assertThat(networkName).isEqualTo("MY_NETWORK");
    }

    @Test
    void sends_transaction_ID_in_proposed_transaction() throws Exception {
        network = gateway.getNetwork("MY_NETWORK");
        Contract contract = network.getContract("CHAINCODE_NAME");

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
        network = gateway.getNetwork("MY_NETWORK");
        Contract contract = network.getContract("CHAINCODE_NAME");
        Proposal.Builder builder = contract.newProposal("TRANSACTION_NAME");

        Proposal proposal1 = builder.build();
        Proposal proposal2 = builder.build();

        assertThat(proposal1.getTransactionId()).isNotEqualTo(proposal2.getTransactionId());
    }

    @Test
    void uses_signer_for_submit() throws Exception {
        Signer signer = (digest) -> "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);
        try (Gateway gateway = mocker.getGatewayBuilder().signer(signer).connect()) {
            network = gateway.getNetwork("NETWORK");

            Contract contract = network.getContract("CHAINCODE_NAME");
            contract.submitTransaction("TRANSACTION_NAME");

            SubmitRequest request = mocker.captureSubmit();
            String signature = request.getPreparedTransaction().getSignature().toStringUtf8();

            assertThat(signature).isEqualTo("MY_SIGNATURE");
        }
    }

    @Test
    void uses_hash() throws Exception {
        List<String> actual = new ArrayList<>();
        Function<byte[], byte[]> hash = (message) -> "MY_DIGEST".getBytes(StandardCharsets.UTF_8);
        Signer signer = (digest) -> {
            actual.add(new String(digest, StandardCharsets.UTF_8));
            return "SIGNATURE".getBytes(StandardCharsets.UTF_8);
        };

        try (Gateway gateway = mocker.getGatewayBuilder().hash(hash).signer(signer).connect()) {
            network = gateway.getNetwork("NETWORK");

            Contract contract = network.getContract("CHAINCODE_NAME");
            contract.submitTransaction("TRANSACTION_NAME");

            assertThat(actual).hasSameElementsAs(Arrays.asList("MY_DIGEST", "MY_DIGEST"));
        }
    }

    @Test
    void throws_on_endorse_connection_error() {
        doThrow(new StatusRuntimeException(io.grpc.Status.UNAVAILABLE)).when(stub).endorse(any());

        Contract contract = network.getContract("CHAINCODE_NAME");

        assertThatThrownBy(() -> contract.submitTransaction("TRANSACTION_NAME"))
                .isInstanceOf(StatusRuntimeException.class);
    }

    @Test
    void throws_on_submit_connection_error() {
        doThrow(new StatusRuntimeException(io.grpc.Status.UNAVAILABLE)).when(stub).submit(any());

        Contract contract = network.getContract("CHAINCODE_NAME");

        assertThatThrownBy(() -> contract.submitTransaction("TRANSACTION_NAME"))
                .isInstanceOf(StatusRuntimeException.class);
    }

    @Test
    void throws_on_commit_failure() {
        doReturn(utils.newCommitStatusResponse(TxValidationCode.MVCC_READ_CONFLICT))
                .when(stub).commitStatus(any());

        Transaction transaction = network.getContract("CHAINCODE_NAME")
                .newProposal("TRANSACTION_NAME")
                .build()
                .endorse();

        CommitException e = catchThrowableOfType(transaction::submit, CommitException.class);

        assertThat(e).hasMessageContaining(TxValidationCode.MVCC_READ_CONFLICT.name());
        assertThat(e.getCode()).isEqualTo(TxValidationCode.MVCC_READ_CONFLICT);
        assertThat(e.getTransactionId()).isEqualTo(transaction.getTransactionId());
    }

    @Test
    void sends_transaction_ID_in_commit_status_request() throws Exception {
        network = gateway.getNetwork("MY_NETWORK");
        Contract contract = network.getContract("CHAINCODE_NAME");
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
        network = gateway.getNetwork("MY_NETWORK");

        Contract contract = network.getContract("CHAINCODE_NAME");
        contract.submitTransaction("TRANSACTION_NAME");

        SignedCommitStatusRequest signedRequest = mocker.captureCommitStatus();
        CommitStatusRequest request = CommitStatusRequest.parseFrom(signedRequest.getRequest());
        String networkName = request.getChannelId();

        assertThat(networkName).isEqualTo("MY_NETWORK");
    }

    @Test
    void commit_returns_transaction_validation_code() {
        doReturn(utils.newCommitStatusResponse(TxValidationCode.MVCC_READ_CONFLICT))
                .when(stub).commitStatus(any());

        Contract contract = network.getContract("CHAINCODE_NAME");
        Status status = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submitAsync()
                .getStatus();

        assertThat(status.getCode()).isEqualTo(TxValidationCode.MVCC_READ_CONFLICT);
    }

    @Test
    void commit_returns_successful_for_successful_transaction() {
        Contract contract = network.getContract("CHAINCODE_NAME");
        Status status = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submitAsync()
                .getStatus();

        assertThat(status.isSuccessful()).isTrue();
    }

    @Test
    void commit_returns_unsuccessful_for_failed_transaction() {
        doReturn(utils.newCommitStatusResponse(TxValidationCode.MVCC_READ_CONFLICT))
                .when(stub).commitStatus(any());

        Contract contract = network.getContract("CHAINCODE_NAME");
        Status status = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submitAsync()
                .getStatus();

        assertThat(status.isSuccessful()).isFalse();
    }

    @Test
    void commit_returns_block_number() {
        doReturn(utils.newCommitStatusResponse(TxValidationCode.MVCC_READ_CONFLICT, 101))
                .when(stub).commitStatus(any());

        Contract contract = network.getContract("CHAINCODE_NAME");
        Status status = contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submitAsync()
                .getStatus();

        assertThat(status.getBlockNumber()).isEqualTo(101);
    }

    @Test
    void endorse_passes_call_options_to_gRPC_client() {
        CallOption expected = CallOption.deadlineAfter(5, TimeUnit.SECONDS);
        Contract contract = network.getContract("MY_CHAINCODE");

        contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse(expected);

        List<CallOption> actual = mocker.captureEndorseOptions();
        assertThat(actual).contains(expected);
    }

    @Test
    void submit_passes_call_options_to_gRPC_client() throws CommitException {
        CallOption expected = CallOption.deadlineAfter(5, TimeUnit.SECONDS);
        Contract contract = network.getContract("MY_CHAINCODE");

        contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submit(expected);

        List<CallOption> actual = mocker.captureSubmitOptions();
        assertThat(actual).contains(expected);
    }

    @Test
    void submit_passes_call_options_to_gRPC_client_commitStatus() throws CommitException {
        CallOption expected = CallOption.deadlineAfter(5, TimeUnit.SECONDS);
        Contract contract = network.getContract("MY_CHAINCODE");

        contract.newProposal("TRANSACTION_NAME")
                .build()
                .endorse()
                .submit(expected);

        List<CallOption> actual = mocker.captureCommitStatusOptions();
        assertThat(actual).contains(expected);
    }
}
