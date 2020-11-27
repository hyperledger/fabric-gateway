/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.stream.Collectors;

import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.Channel;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.ContractException;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.GatewayServiceStub;
import org.hyperledger.fabric.client.MockGatewayService;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.TestUtils;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.gateway.PreparedTransaction;
import org.hyperledger.fabric.gateway.ProposedTransaction;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.peer.Chaincode;
import org.hyperledger.fabric.protos.peer.ProposalPackage;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;
import org.mockito.Captor;
import org.mockito.Mockito;
import org.mockito.MockitoSession;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.spy;

public class ContractTest {
    private static final TestUtils utils = TestUtils.getInstance();
    private static final Gateway.Builder builder = utils.newGatewayBuilder();
    private static final GatewayServiceStub STUB = new GatewayServiceStub();

    private GatewayServiceStub stub;
    private Gateway gateway;
    private Network network;

    private MockitoSession mockitoSession;
    @Captor
    private ArgumentCaptor<ProposedTransaction> proposedTransactionCaptor;
    @Captor
    private ArgumentCaptor<PreparedTransaction> preparedTransactionCaptor;

    private Chaincode.ChaincodeSpec getChaincodeSpec() throws InvalidProtocolBufferException {
        ProposalPackage.Proposal proposal = getProposal();
        ProposalPackage.ChaincodeProposalPayload chaincodeProposalPayload = ProposalPackage.ChaincodeProposalPayload.parseFrom(proposal.getPayload());
        Chaincode.ChaincodeInvocationSpec chaincodeInvocationSpec = Chaincode.ChaincodeInvocationSpec.parseFrom(chaincodeProposalPayload.getInput());
        return chaincodeInvocationSpec.getChaincodeSpec();
    }

    private ProposalPackage.Proposal getProposal() throws InvalidProtocolBufferException {
        ProposedTransaction request = proposedTransactionCaptor.getValue();
        return ProposalPackage.Proposal.parseFrom(request.getProposal().getProposalBytes());
    }

    private Common.SignatureHeader getSignatureHeader() throws InvalidProtocolBufferException {
        Common.Header header = getHeader();
        return Common.SignatureHeader.parseFrom(header.getSignatureHeader());
    }

    private Common.ChannelHeader getChannelHeader() throws InvalidProtocolBufferException {
        Common.Header header = getHeader();
        return Common.ChannelHeader.parseFrom(header.getChannelHeader());
    }

    private Common.Header getHeader() throws InvalidProtocolBufferException {
        ProposalPackage.Proposal proposal = getProposal();
        return Common.Header.parseFrom(proposal.getHeader());
    }

    @BeforeEach
    void beforeEach() {
        mockitoSession = Mockito.mockitoSession()
                .initMocks(this)
                .startMocking();

        stub = spy(STUB);
        MockGatewayService service = new MockGatewayService(stub);
        Channel channel = utils.newChannelForService(service);
        gateway = builder.connection(channel).connect();
        network = gateway.getNetwork("NETWORK");
    }

    @AfterEach
    void afterEach() {
        gateway.close();
        mockitoSession.finishMocking();
    }

    @Test
    public void evaluateTransaction_returns_gateway_response() throws ContractException {
        doReturn(utils.newResult("MY_RESULT"))
                .when(stub).evaluate(any());

        Contract contract = network.getContract("CHAINCODE_ID");
        byte[] actual = contract.evaluateTransaction("TRANSACTION_NAME");

        assertThat(actual).asString(StandardCharsets.UTF_8).isEqualTo("MY_RESULT");
    }

    @Test
    public void evaluateTransaction_sends_chaincode_ID() throws ContractException, InvalidProtocolBufferException {
        Contract contract = network.getContract("MY_CHAINCODE_ID");
        contract.evaluateTransaction("TRANSACTION_NAME");

        Mockito.verify(stub).evaluate(proposedTransactionCaptor.capture());
        String actual = getChaincodeSpec().getChaincodeId().getName();

        assertThat(actual).isEqualTo("MY_CHAINCODE_ID");
    }

    @Test
    public void evaluateTransaction_sends_transaction_name_for_default_contract() throws ContractException, InvalidProtocolBufferException {
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.evaluateTransaction("MY_TRANSACTION_NAME");

        Mockito.verify(stub).evaluate(proposedTransactionCaptor.capture());
        List<String> chaincodeArgs = getChaincodeSpec().getInput().getArgsList().stream()
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).first().isEqualTo("MY_TRANSACTION_NAME");
    }

    @Test
    public void evaluateTransaction_sends_transaction_name_for_specified_contract() throws ContractException, InvalidProtocolBufferException {
        Contract contract = network.getContract("CHAINCODE_ID", "MY_CONTRACT");
        contract.evaluateTransaction("MY_TRANSACTION_NAME");

        Mockito.verify(stub).evaluate(proposedTransactionCaptor.capture());
        List<String> chaincodeArgs = getChaincodeSpec().getInput().getArgsList().stream()
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).first().isEqualTo("MY_CONTRACT:MY_TRANSACTION_NAME");
    }

    @Test
    public void evaluateTransaction_sends_transaction_arguments() throws ContractException, InvalidProtocolBufferException {
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.evaluateTransaction("TRANSACTION_NAME", "one", "two", "three");

        Mockito.verify(stub).evaluate(proposedTransactionCaptor.capture());
        List<String> chaincodeArgs = getChaincodeSpec().getInput().getArgsList().stream()
                .skip(1)
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).containsExactly("one", "two", "three");
    }

    @Test
    public void evaluateTransaction_uses_signer() throws ContractException {
        Signer signer = (digest) -> "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);
        gateway = builder.signer(signer).connect();

        network = gateway.getNetwork("NETWORK");
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.evaluateTransaction("TRANSACTION_NAME");

        Mockito.verify(stub).evaluate(proposedTransactionCaptor.capture());
        ProposedTransaction request = proposedTransactionCaptor.getValue();
        String signature = request.getProposal().getSignature().toStringUtf8();

        assertThat(signature).isEqualTo("MY_SIGNATURE");
    }

    @Test
    public void evaluateTransaction_uses_identity() throws ContractException, InvalidProtocolBufferException {
        Identity identity = new X509Identity("MY_MSP_ID", utils.getCredentials().getCertificate());
        gateway = builder.identity(identity).connect();

        network = gateway.getNetwork("NETWORK");
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.evaluateTransaction("TRANSACTION_NAME");

        Mockito.verify(stub).evaluate(proposedTransactionCaptor.capture());
        ByteString serializedIdentity = getSignatureHeader().getCreator();

        byte[] expected = GatewayUtils.serializeIdentity(identity);
        assertThat(serializedIdentity.toByteArray()).isEqualTo(expected);
    }

    @Test
    public void evaluateTransaction_sends_network_name() throws ContractException, InvalidProtocolBufferException {
        network = gateway.getNetwork("MY_NETWORK");
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.evaluateTransaction("TRANSACTION_NAME");

        Mockito.verify(stub).evaluate(proposedTransactionCaptor.capture());
        String networkName = getChannelHeader().getChannelId();

        assertThat(networkName).isEqualTo("MY_NETWORK");
    }

    @Test
    public void submitTransaction_returns_gateway_response() throws Exception {
        doReturn(utils.newPreparedTransaction("MY_RESULT", "SIGNATURE"))
                .when(stub).endorse(any());

        Contract contract = network.getContract("CHAINCODE_ID");
        byte[] actual = contract.submitTransaction("TRANSACTION_NAME");

        assertThat(actual).asString(StandardCharsets.UTF_8).isEqualTo("MY_RESULT");
    }

    @Test
    public void submitTransaction_sends_chaincode_ID() throws Exception {
        Contract contract = network.getContract("MY_CHAINCODE_ID");
        contract.submitTransaction("TRANSACTION_NAME");

        Mockito.verify(stub).endorse(proposedTransactionCaptor.capture());
        String actual = getChaincodeSpec().getChaincodeId().getName();

        assertThat(actual).isEqualTo("MY_CHAINCODE_ID");
    }

    @Test
    public void submitTransaction_sends_transaction_name_for_default_contract() throws Exception {
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.submitTransaction("MY_TRANSACTION_NAME");

        Mockito.verify(stub).endorse(proposedTransactionCaptor.capture());
        List<String> chaincodeArgs = getChaincodeSpec().getInput().getArgsList().stream()
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).first().isEqualTo("MY_TRANSACTION_NAME");
    }

    @Test
    public void submitTransaction_sends_transaction_name_for_specified_contract() throws Exception {
        Contract contract = network.getContract("CHAINCODE_ID", "MY_CONTRACT");
        contract.submitTransaction("MY_TRANSACTION_NAME");

        Mockito.verify(stub).endorse(proposedTransactionCaptor.capture());
        List<String> chaincodeArgs = getChaincodeSpec().getInput().getArgsList().stream()
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).first().isEqualTo("MY_CONTRACT:MY_TRANSACTION_NAME");
    }

    @Test
    public void submitTransaction_sends_transaction_arguments() throws Exception {
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.submitTransaction("TRANSACTION_NAME", "one", "two", "three");

        Mockito.verify(stub).endorse(proposedTransactionCaptor.capture());
        List<String> chaincodeArgs = getChaincodeSpec().getInput().getArgsList().stream()
                .skip(1)
                .map(ByteString::toStringUtf8)
                .collect(Collectors.toList());

        assertThat(chaincodeArgs).containsExactly("one", "two", "three");
    }

    @Test
    public void submitTransaction_uses_signer_for_endorse() throws Exception {
        Signer signer = (digest) -> "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);
        gateway = builder.signer(signer).connect();

        network = gateway.getNetwork("NETWORK");
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.submitTransaction("TRANSACTION_NAME");

        Mockito.verify(stub).endorse(proposedTransactionCaptor.capture());
        ProposedTransaction request = proposedTransactionCaptor.getValue();
        String signature = request.getProposal().getSignature().toStringUtf8();

        assertThat(signature).isEqualTo("MY_SIGNATURE");
    }

    @Test
    public void submitTransaction_uses_identity() throws Exception {
        Identity identity = new X509Identity("MY_MSP_ID", utils.getCredentials().getCertificate());
        gateway = builder.identity(identity).connect();

        network = gateway.getNetwork("NETWORK");
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.submitTransaction("TRANSACTION_NAME");

        Mockito.verify(stub).endorse(proposedTransactionCaptor.capture());
        ByteString serializedIdentity = getSignatureHeader().getCreator();

        byte[] expected = GatewayUtils.serializeIdentity(identity);
        assertThat(serializedIdentity.toByteArray()).isEqualTo(expected);
    }

    @Test
    public void submitTransaction_sends_network_name() throws Exception {
        network = gateway.getNetwork("MY_NETWORK");
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.submitTransaction("TRANSACTION_NAME");

        Mockito.verify(stub).endorse(proposedTransactionCaptor.capture());
        String networkName = getChannelHeader().getChannelId();

        assertThat(networkName).isEqualTo("MY_NETWORK");
    }

    @Test
    public void submitTransaction_uses_signer_for_submit() throws Exception {
        Signer signer = (digest) -> "MY_SIGNATURE".getBytes(StandardCharsets.UTF_8);
        gateway = builder.signer(signer).connect();

        network = gateway.getNetwork("NETWORK");
        Contract contract = network.getContract("CHAINCODE_ID");
        contract.submitTransaction("TRANSACTION_NAME");

        Mockito.verify(stub).submit(preparedTransactionCaptor.capture());
        PreparedTransaction request = preparedTransactionCaptor.getValue();
        String signature = request.getEnvelope().getSignature().toStringUtf8();

        assertThat(signature).isEqualTo("MY_SIGNATURE");
    }
}
