/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.CallOptions;
import io.grpc.Deadline;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.protos.gateway.EvaluateRequest;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;
import java.util.function.Function;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.Mockito.*;

public final class EvaluateTransactionTest {
    private static final TestUtils utils = TestUtils.getInstance();
    private static final Deadline defaultDeadline = Deadline.after(1, TimeUnit.DAYS);

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
                .evaluateOptions(options -> options.withDeadline(defaultDeadline))
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
        assertThatThrownBy(() -> contract.evaluateTransaction(null))
                .isInstanceOf(NullPointerException.class)
                .hasMessageContaining("transaction name");
    }

    @Test
    void returns_gateway_response() throws GatewayException {
        doReturn(utils.newEvaluateResponse("MY_RESULT"))
                .when(stub).evaluate(any());

        byte[] actual = contract.evaluateTransaction("TRANSACTION_NAME");

        assertThat(actual).asString(StandardCharsets.UTF_8).isEqualTo("MY_RESULT");
    }

    @Test
    void sends_chaincode_name() throws InvalidProtocolBufferException, GatewayException {
        contract.evaluateTransaction("TRANSACTION_NAME");

        EvaluateRequest request = mocker.captureEvaluate();
        String actual = mocker.getChaincodeSpec(request.getProposedTransaction()).getChaincodeId().getName();

        assertThat(actual).isEqualTo(contract.getChaincodeName());
    }

    @Test
    void sends_transaction_name_for_default_contract() throws InvalidProtocolBufferException, GatewayException {
        network.getContract("CHAINCODE_NAME")
                .evaluateTransaction("MY_TRANSACTION_NAME");

        EvaluateRequest request = mocker.captureEvaluate();
        List<String> chaincodeArgs = mocker.getChaincodeSpec(request.getProposedTransaction()).getInput().getArgsList().stream()
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).first().isEqualTo("MY_TRANSACTION_NAME");
    }

    @Test
    void sends_transaction_name_for_specified_contract() throws InvalidProtocolBufferException, GatewayException {
        network.getContract("CHAINCODE_NAME", "MY_CONTRACT")
                .evaluateTransaction("MY_TRANSACTION_NAME");

        EvaluateRequest request = mocker.captureEvaluate();
        List<String> chaincodeArgs = mocker.getChaincodeSpec(request.getProposedTransaction()).getInput().getArgsList().stream()
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).first().isEqualTo("MY_CONTRACT:MY_TRANSACTION_NAME");
    }

    @Test
    void sends_transaction_string_arguments() throws InvalidProtocolBufferException, GatewayException {
        contract.evaluateTransaction("TRANSACTION_NAME", "one", "two", "three");

        EvaluateRequest request = mocker.captureEvaluate();
        List<String> chaincodeArgs = mocker.getChaincodeSpec(request.getProposedTransaction()).getInput().getArgsList().stream()
                .skip(1)
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).containsExactly("one", "two", "three");
    }

    @Test
    void sends_transaction_byte_array_arguments() throws InvalidProtocolBufferException, GatewayException {
        byte[][] arguments = Stream.of("one", "two", "three")
                .map(s -> s.getBytes(StandardCharsets.UTF_8))
                .toArray(byte[][]::new);
        contract.evaluateTransaction("TRANSACTION_NAME", arguments);

        EvaluateRequest request = mocker.captureEvaluate();
        byte[][] chaincodeArgs = mocker.getChaincodeSpec(request.getProposedTransaction()).getInput().getArgsList().stream()
                .skip(1)
                .map(ByteString::toByteArray)
                .toArray(byte[][]::new);

        assertThat(chaincodeArgs).isDeepEqualTo(arguments);
    }

    @Test
    void uses_signer() throws GatewayException {
        Signer signer = (digest) -> "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);
        try (Gateway gateway = mocker.getGatewayBuilder().signer(signer).connect()) {
            gateway.getNetwork("NETWORK")
                    .getContract("CHAINCODE_NAME")
                    .evaluateTransaction("TRANSACTION_NAME");

            EvaluateRequest request = mocker.captureEvaluate();
            String signature = request.getProposedTransaction().getSignature().toStringUtf8();

            assertThat(signature).isEqualTo("MY_SIGNATURE");
        }
    }

    @Test
    void uses_hash() throws GatewayException {
        AtomicReference<String> actual = new AtomicReference<>();
        Function<byte[], byte[]> hash = message -> "MY_DIGEST".getBytes(StandardCharsets.UTF_8);
        Signer signer = digest -> {
            actual.set(new String(digest, StandardCharsets.UTF_8));
            return "SIGNATURE".getBytes(StandardCharsets.UTF_8);
        };

        try (Gateway gateway = mocker.getGatewayBuilder().hash(hash).signer(signer).connect()) {
            gateway.getNetwork("NETWORK")
                    .getContract("CHAINCODE_NAME")
                    .evaluateTransaction("TRANSACTION_NAME");

            assertThat(actual.get()).isEqualTo("MY_DIGEST");
        }
    }

    @Test
    void throws_on_connection_error() {
        doThrow(new StatusRuntimeException(Status.UNAVAILABLE)).when(stub).evaluate(any());

        assertThatThrownBy(() -> contract.evaluateTransaction("TRANSACTION_NAME"))
                .isInstanceOf(GatewayException.class)
                .extracting(t -> ((GatewayException) t).getStatus())
                .isEqualTo(io.grpc.Status.UNAVAILABLE);
    }

    @Test
    void uses_identity() throws Exception {
        Identity identity = new X509Identity("MY_MSP_ID", utils.getCredentials().getCertificate());
        try (Gateway gateway = mocker.getGatewayBuilder().identity(identity).connect()) {
            gateway.getNetwork("NETWORK")
                    .getContract("CHAINCODE_NAME")
                    .evaluateTransaction("TRANSACTION_NAME");

            EvaluateRequest request = mocker.captureEvaluate();
            ByteString serializedIdentity = mocker.getSignatureHeader(request.getProposedTransaction()).getCreator();

            byte[] expected = GatewayUtils.serializeIdentity(identity);
            assertThat(serializedIdentity.toByteArray()).isEqualTo(expected);
        }
    }

    @Test
    void sends_network_name_in_proposal_for_default_contract() throws InvalidProtocolBufferException, GatewayException {
        network.getContract("CHAINCODE_NAME")
                .evaluateTransaction("TRANSACTION_NAME");

        EvaluateRequest request = mocker.captureEvaluate();
        String networkName = mocker.getChannelHeader(request.getProposedTransaction()).getChannelId();

        assertThat(networkName).isEqualTo(network.getName());
    }

    @Test
    void sends_network_name_in_proposed_transaction_for_default_contract() throws GatewayException {
        network.getContract("CHAINCODE_NAME")
                .evaluateTransaction("TRANSACTION_NAME");

        EvaluateRequest request = mocker.captureEvaluate();
        String networkName = request.getChannelId();

        assertThat(networkName).isEqualTo(network.getName());
    }

    @Test
    void sends_network_name_in_proposal_for_specified_contract() throws InvalidProtocolBufferException, GatewayException {
        network.getContract("CHAINCODE_NAME", "CONTRACT_NAME")
                .evaluateTransaction("TRANSACTION_NAME");

        EvaluateRequest request = mocker.captureEvaluate();
        String networkName = mocker.getChannelHeader(request.getProposedTransaction()).getChannelId();

        assertThat(networkName).isEqualTo(network.getName());
    }

    @Test
    void sends_network_name_in_proposed_transaction_for_specified_contract() throws GatewayException {
        network.getContract("CHAINCODE_NAME", "CONTRACT_NAME")
                .evaluateTransaction("TRANSACTION_NAME");

        EvaluateRequest request = mocker.captureEvaluate();
        String networkName = request.getChannelId();

        assertThat(networkName).isEqualTo(network.getName());
    }

    @Test
    void sends_transaction_ID_in_proposed_transaction() throws GatewayException {
        Proposal proposal = contract.newProposal("TRANSACTION_NAME").build();
        proposal.evaluate();

        String expected = proposal.getTransactionId();
        assertThat(expected).isNotEmpty();

        EvaluateRequest request = mocker.captureEvaluate();
        String actual = request.getTransactionId();

        assertThat(actual).isEqualTo(expected);
    }

    @Test
    void sends_byte_array_transient_data() throws Exception {
        contract.newProposal("TRANSACTION_NAME")
                .putTransient("uno", "one".getBytes(StandardCharsets.UTF_8))
                .putTransient("dos", "two".getBytes(StandardCharsets.UTF_8))
                .build()
                .evaluate();

        EvaluateRequest request = mocker.captureEvaluate();
        Map<String, ByteString> transientData = mocker.getProposalPayload(request.getProposedTransaction()).getTransientMapMap();
        assertThat(transientData).containsOnly(
                entry("uno", ByteString.copyFromUtf8("one")),
                entry("dos", ByteString.copyFromUtf8("two")));
    }

    @Test
    void sends_string_transient_data() throws Exception {
        contract.newProposal("TRANSACTION_NAME")
                .putTransient("uno", "one")
                .putTransient("dos", "two")
                .build()
                .evaluate();

        EvaluateRequest request = mocker.captureEvaluate();
        Map<String, ByteString> transientData = mocker.getProposalPayload(request.getProposedTransaction()).getTransientMapMap();
        assertThat(transientData).containsOnly(
                entry("uno", ByteString.copyFromUtf8("one")),
                entry("dos", ByteString.copyFromUtf8("two")));
    }

    @Test
    void sets_endorsing_orgs() throws GatewayException {
        contract.newProposal("TRANSACTION_NAME")
                .setEndorsingOrganizations("Org1MSP", "Org3MSP")
                .build()
                .evaluate();

        EvaluateRequest request = mocker.captureEvaluate();
        List<String> endorsingOrgs = request.getTargetOrganizationsList();
        assertThat(endorsingOrgs).containsExactlyInAnyOrder("Org1MSP", "Org3MSP");
    }

    @Test
    @SuppressWarnings("deprecation")
    void uses_legacy_specified_call_options() throws GatewayException {
        Deadline expected = Deadline.after(1, TimeUnit.MINUTES);

        contract.newProposal("TRANSACTION_NAME")
                .build()
                .evaluate(CallOption.deadline(expected));

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual).first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(expected);
    }

    @Test
    void uses_specified_call_options() throws GatewayException {
        Deadline expected = Deadline.after(1, TimeUnit.MINUTES);

        contract.newProposal("TRANSACTION_NAME")
                .build()
                .evaluate(callOptions -> callOptions.withDeadline(expected));

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual).first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(expected);
    }

    @Test
    @SuppressWarnings("deprecation")
    void uses_legacy_default_call_options() throws GatewayException {
        Deadline expected = Deadline.after(1, TimeUnit.MINUTES);

        try (Gateway gateway = mocker.getGatewayBuilder()
                .evaluateOptions(CallOption.deadline(expected))
                .connect()) {
            gateway.getNetwork("NETWORK")
                    .getContract("CHAINCODE_NAME")
                    .newProposal("TRANSACTION_NAME")
                    .build()
                    .evaluate();
        }

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual).first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(expected);
    }

    @Test
    void uses_default_call_options() throws GatewayException {
        contract.newProposal("TRANSACTION_NAME")
                .build()
                .evaluate();

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual).first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(defaultDeadline);
    }
}
