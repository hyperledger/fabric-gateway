/*
 *  Copyright 2020 IBM All Rights Reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.List;
import java.util.concurrent.TimeUnit;
import java.util.stream.Stream;

import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.CallOptions;
import io.grpc.ManagedChannel;
import org.hyperledger.fabric.protos.common.ChannelHeader;
import org.hyperledger.fabric.protos.common.Envelope;
import org.hyperledger.fabric.protos.common.Header;
import org.hyperledger.fabric.protos.common.SignatureHeader;
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.EvaluateRequest;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.hyperledger.fabric.protos.peer.ChaincodeInvocationSpec;
import org.hyperledger.fabric.protos.peer.ChaincodeProposalPayload;
import org.hyperledger.fabric.protos.peer.ChaincodeSpec;
import org.hyperledger.fabric.protos.peer.SignedProposal;
import org.mockito.ArgumentCaptor;
import org.mockito.Captor;
import org.mockito.Mockito;
import org.mockito.MockitoSession;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.atLeastOnce;
import static org.mockito.Mockito.spy;

/**
 * Registers mock Gateway and Deliver gRPC services with a real gRPC channel, and connects a Gateway instance to that
 * channel, for use in unit tests. Provides access to the Gateway and Deliver service implementation stubs so that
 * mock return values can be specified, and methods to capture request parameters passed to gRPC service methods.
 */
public final class GatewayMocker implements AutoCloseable {
    private static final TestUtils utils = TestUtils.getInstance();

    private final GatewayServiceStub gatewaySpy;
    private final DeliverServiceStub deliverSpy;
    private final ManagedChannel channelSpy;
    private final Gateway.Builder builder;

    private final MockitoSession mockitoSession;
    @Captor private ArgumentCaptor<EndorseRequest> endorseRequestCaptor;
    @Captor private ArgumentCaptor<EvaluateRequest> evaluateRequestCaptor;
    @Captor private ArgumentCaptor<SubmitRequest> submitRequestCaptor;
    @Captor private ArgumentCaptor<SignedCommitStatusRequest> commitStatusRequestCaptor;
    @Captor private ArgumentCaptor<SignedChaincodeEventsRequest> chaincodeEventsRequestCaptor;
    @Captor private ArgumentCaptor<Stream<Envelope>> deliverRequestCaptor;
    @Captor private ArgumentCaptor<Stream<Envelope>> deliverFilteredRequestCaptor;
    @Captor private ArgumentCaptor<Stream<Envelope>> deliverWithPrivateDataRequestCaptor;
    @Captor private ArgumentCaptor<CallOptions> callOptionsCaptor;

    public GatewayMocker() {
        this(utils.newGatewayBuilder());
    }

    public GatewayMocker(final Gateway.Builder builder) {
        this.builder = builder;
        mockitoSession = Mockito.mockitoSession()
                .initMocks(this)
                .startMocking();

        gatewaySpy = spy(new GatewayServiceStub());
        MockGatewayService gatewayService = new MockGatewayService(gatewaySpy);

        deliverSpy = spy(new DeliverServiceStub());
        MockDeliverService deliverService = new MockDeliverService(deliverSpy);

        channelSpy = spy(new WrappedManagedChannel(utils.newChannelForServices(gatewayService, deliverService)));
        builder.connection(channelSpy);
    }

    /**
     * Reset stubs/spies.
     */
    public void reset() {
        Mockito.reset(gatewaySpy, channelSpy);
    }

    public void close() {
        utils.shutdownChannel(channelSpy, 5, TimeUnit.SECONDS);
        mockitoSession.finishMocking();
    }

    public GatewayServiceStub getGatewayServiceStubSpy() {
        return gatewaySpy;
    }

    public DeliverServiceStub getDeliverServiceStubSpy() {
        return deliverSpy;
    }

    public Gateway.Builder getGatewayBuilder() {
        return builder;
    }

    public EndorseRequest captureEndorse() {
        Mockito.verify(gatewaySpy).endorse(endorseRequestCaptor.capture());
        return endorseRequestCaptor.getValue();
    }

    public EvaluateRequest captureEvaluate() {
        Mockito.verify(gatewaySpy).evaluate(evaluateRequestCaptor.capture());
        return evaluateRequestCaptor.getValue();
    }

    public SubmitRequest captureSubmit() {
        Mockito.verify(gatewaySpy).submit(submitRequestCaptor.capture());
        return submitRequestCaptor.getValue();
    }

    public SignedCommitStatusRequest captureCommitStatus() {
        Mockito.verify(gatewaySpy).commitStatus(commitStatusRequestCaptor.capture());
        return commitStatusRequestCaptor.getValue();
    }

    public SignedChaincodeEventsRequest captureChaincodeEvents() {
        Mockito.verify(gatewaySpy).chaincodeEvents(chaincodeEventsRequestCaptor.capture());
        return chaincodeEventsRequestCaptor.getValue();
    }

    public Stream<Envelope> captureBlockEvents() {
        Mockito.verify(deliverSpy).blockEvents(deliverRequestCaptor.capture());
        return deliverRequestCaptor.getValue();
    }

    public Stream<Envelope> captureFilteredBlockEvents() {
        Mockito.verify(deliverSpy).filteredBlockEvents(deliverFilteredRequestCaptor.capture());
        return deliverFilteredRequestCaptor.getValue();
    }

    public Stream<Envelope> captureBlockAndPrivateDataEvents() {
        Mockito.verify(deliverSpy).blockAndPrivateDataEvents(deliverWithPrivateDataRequestCaptor.capture());
        return deliverWithPrivateDataRequestCaptor.getValue();
    }

    public List<CallOptions> captureCallOptions() {
        Mockito.verify(channelSpy, atLeastOnce()).newCall(any(), callOptionsCaptor.capture());
        return callOptionsCaptor.getAllValues();
    }

    public ChaincodeSpec getChaincodeSpec(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        org.hyperledger.fabric.protos.peer.Proposal proposal = getProposal(proposedTransaction);
        ChaincodeProposalPayload chaincodeProposalPayload = ChaincodeProposalPayload.parseFrom(proposal.getPayload());
        ChaincodeInvocationSpec chaincodeInvocationSpec = ChaincodeInvocationSpec.parseFrom(chaincodeProposalPayload.getInput());
        return chaincodeInvocationSpec.getChaincodeSpec();
    }

    public ChaincodeProposalPayload getProposalPayload(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        org.hyperledger.fabric.protos.peer.Proposal proposal = getProposal(proposedTransaction);
        return ChaincodeProposalPayload.parseFrom(proposal.getPayload());
    }

    public org.hyperledger.fabric.protos.peer.Proposal getProposal(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        return org.hyperledger.fabric.protos.peer.Proposal.parseFrom(proposedTransaction.getProposalBytes());
    }

    public SignatureHeader getSignatureHeader(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        Header header = getHeader(proposedTransaction);
        return SignatureHeader.parseFrom(header.getSignatureHeader());
    }

    public ChannelHeader getChannelHeader(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        Header header = getHeader(proposedTransaction);
        return ChannelHeader.parseFrom(header.getChannelHeader());
    }

    public Header getHeader(SignedProposal proposedTransaction) throws InvalidProtocolBufferException {
        org.hyperledger.fabric.protos.peer.Proposal proposal = getProposal(proposedTransaction);
        return Header.parseFrom(proposal.getHeader());
    }
}
