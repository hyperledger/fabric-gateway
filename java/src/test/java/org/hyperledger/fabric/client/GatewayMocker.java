/*
 *  Copyright 2020 IBM All Rights Reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.concurrent.TimeUnit;

import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.ManagedChannel;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.EvaluateRequest;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.hyperledger.fabric.protos.peer.Chaincode;
import org.hyperledger.fabric.protos.peer.ProposalPackage;
import org.hyperledger.fabric.protos.peer.ProposalPackage.SignedProposal;
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
    private ArgumentCaptor<EndorseRequest> endorseRequestCaptor;
    @Captor
    private ArgumentCaptor<EvaluateRequest> evaluateRequestCaptor;
    @Captor
    private ArgumentCaptor<SubmitRequest> submitRequestCaptor;
    @Captor
    private ArgumentCaptor<SignedCommitStatusRequest> commitStatusRequestCaptor;
    @Captor
    private ArgumentCaptor<SignedChaincodeEventsRequest> chaincodeEventsRequestCaptor;

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

    public EndorseRequest captureEndorse() {
        Mockito.verify(stub).endorse(endorseRequestCaptor.capture());
        return endorseRequestCaptor.getValue();
    }

    public EvaluateRequest captureEvaluate() {
        Mockito.verify(stub).evaluate(evaluateRequestCaptor.capture());
        return evaluateRequestCaptor.getValue();
    }

    public SubmitRequest captureSubmit() {
        Mockito.verify(stub).submit(submitRequestCaptor.capture());
        return submitRequestCaptor.getValue();
    }

    public SignedCommitStatusRequest captureCommitStatus() {
        Mockito.verify(stub).commitStatus(commitStatusRequestCaptor.capture());
        return commitStatusRequestCaptor.getValue();
    }

    public SignedChaincodeEventsRequest captureChaincodeEvents() {
        Mockito.verify(stub).chaincodeEvents(chaincodeEventsRequestCaptor.capture());
        return chaincodeEventsRequestCaptor.getValue();
    }

    public Chaincode.ChaincodeSpec getChaincodeSpec(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        ProposalPackage.Proposal proposal = getProposal(proposedTransaction);
        ProposalPackage.ChaincodeProposalPayload chaincodeProposalPayload = ProposalPackage.ChaincodeProposalPayload.parseFrom(proposal.getPayload());
        Chaincode.ChaincodeInvocationSpec chaincodeInvocationSpec = Chaincode.ChaincodeInvocationSpec.parseFrom(chaincodeProposalPayload.getInput());
        return chaincodeInvocationSpec.getChaincodeSpec();
    }

    public ProposalPackage.ChaincodeProposalPayload getProposalPayload(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        ProposalPackage.Proposal proposal = getProposal(proposedTransaction);
        ProposalPackage.ChaincodeProposalPayload chaincodeProposalPayload = ProposalPackage.ChaincodeProposalPayload.parseFrom(proposal.getPayload());
        return chaincodeProposalPayload;
    }

    public ProposalPackage.Proposal getProposal(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        return ProposalPackage.Proposal.parseFrom(proposedTransaction.getProposalBytes());
    }

    public Common.SignatureHeader getSignatureHeader(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        Common.Header header = getHeader(proposedTransaction);
        return Common.SignatureHeader.parseFrom(header.getSignatureHeader());
    }

    public Common.ChannelHeader getChannelHeader(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        Common.Header header = getHeader(proposedTransaction);
        return Common.ChannelHeader.parseFrom(header.getChannelHeader());
    }

    public Common.Header getHeader(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        ProposalPackage.Proposal proposal = getProposal(proposedTransaction);
        return Common.Header.parseFrom(proposal.getHeader());
    }
}
