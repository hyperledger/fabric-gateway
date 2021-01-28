/*
 *  Copyright 2020 IBM All Rights Reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.concurrent.TimeUnit;

import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.ManagedChannel;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.ProposedTransaction;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.peer.Chaincode;
import org.hyperledger.fabric.protos.peer.ProposalPackage;
import org.mockito.ArgumentCaptor;
import org.mockito.Captor;
import org.mockito.Mockito;
import org.mockito.MockitoSession;

import static org.mockito.Mockito.spy;

public final class GatewayMocker implements AutoCloseable {
    private static final TestUtils utils = TestUtils.getInstance();
    private static final GatewayServiceStub STUB = new GatewayServiceStub();

    private final GatewayServiceStub stub;
    private final ManagedChannel channel;
    private final Gateway.Builder builder;

    private final MockitoSession mockitoSession;
    @Captor
    private ArgumentCaptor<ProposedTransaction> proposedTransactionCaptor;
    @Captor
    private ArgumentCaptor<PreparedTransaction> preparedTransactionCaptor;

    public GatewayMocker() {
        this(utils.newGatewayBuilder());
    }

    public GatewayMocker(final Gateway.Builder builder) {
        this.builder = builder;
        mockitoSession = Mockito.mockitoSession()
                .initMocks(this)
                .startMocking();

        stub = spy(STUB);
        MockGatewayService service = new MockGatewayService(stub);
        channel = utils.newChannelForService(service);
        builder.connection(channel);
    }

    public void close() {
        GatewayUtils.shutdownChannel(channel, 5, TimeUnit.SECONDS);
        mockitoSession.finishMocking();
    }

    public GatewayServiceStub getServiceStubSpy() {
        return stub;
    }

    public Gateway.Builder getGatewayBuilder() {
        return builder;
    }

    public ProposedTransaction captureEndorse() {
        Mockito.verify(stub).endorse(proposedTransactionCaptor.capture());
        return proposedTransactionCaptor.getValue();
    }

    public ProposedTransaction captureEvaluate() {
        Mockito.verify(stub).evaluate(proposedTransactionCaptor.capture());
        return proposedTransactionCaptor.getValue();
    }

    public PreparedTransaction captureSubmit() {
        Mockito.verify(stub).submit(preparedTransactionCaptor.capture());
        return preparedTransactionCaptor.getValue();
    }

    public Chaincode.ChaincodeSpec getChaincodeSpec(ProposedTransaction request) throws InvalidProtocolBufferException {
        ProposalPackage.Proposal proposal = getProposal(request);
        ProposalPackage.ChaincodeProposalPayload chaincodeProposalPayload = ProposalPackage.ChaincodeProposalPayload.parseFrom(proposal.getPayload());
        Chaincode.ChaincodeInvocationSpec chaincodeInvocationSpec = Chaincode.ChaincodeInvocationSpec.parseFrom(chaincodeProposalPayload.getInput());
        return chaincodeInvocationSpec.getChaincodeSpec();
    }

    public ProposalPackage.Proposal getProposal(ProposedTransaction request) throws InvalidProtocolBufferException {
        return ProposalPackage.Proposal.parseFrom(request.getProposal().getProposalBytes());
    }

    public Common.SignatureHeader getSignatureHeader(ProposedTransaction request) throws InvalidProtocolBufferException {
        Common.Header header = getHeader(request);
        return Common.SignatureHeader.parseFrom(header.getSignatureHeader());
    }

    public Common.ChannelHeader getChannelHeader(ProposedTransaction request) throws InvalidProtocolBufferException {
        Common.Header header = getHeader(request);
        return Common.ChannelHeader.parseFrom(header.getChannelHeader());
    }

    public Common.Header getHeader(ProposedTransaction request) throws InvalidProtocolBufferException {
        ProposalPackage.Proposal proposal = getProposal(request);
        return Common.Header.parseFrom(proposal.getHeader());
    }
}
