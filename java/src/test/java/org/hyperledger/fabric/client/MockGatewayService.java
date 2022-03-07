/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.stub.StreamObserver;
import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;
import org.hyperledger.fabric.protos.gateway.CommitStatusResponse;
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.EndorseResponse;
import org.hyperledger.fabric.protos.gateway.EvaluateRequest;
import org.hyperledger.fabric.protos.gateway.EvaluateResponse;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.hyperledger.fabric.protos.gateway.SubmitResponse;

/**
 * Mock Gateway gRPC service that acts as an adapter for a stub implementation.
 */
public final class MockGatewayService extends GatewayGrpc.GatewayImplBase {
    private static final TestUtils testUtils = TestUtils.getInstance();

    private final GatewayServiceStub stub;

    public MockGatewayService(final GatewayServiceStub stub) {
        this.stub = stub;
    }

    @Override
    public void endorse(final EndorseRequest request, final StreamObserver<EndorseResponse> responseObserver) {
        testUtils.invokeStubUnaryCall(stub::endorse, request, responseObserver);
    }

    @Override
    public void submit(final SubmitRequest request, final StreamObserver<SubmitResponse> responseObserver) {
        testUtils.invokeStubUnaryCall(stub::submit, request, responseObserver);
    }

    @Override
    public void evaluate(final EvaluateRequest request, final StreamObserver<EvaluateResponse> responseObserver) {
        testUtils.invokeStubUnaryCall(stub::evaluate, request, responseObserver);
    }

    @Override
    public void commitStatus(final SignedCommitStatusRequest request, final StreamObserver<CommitStatusResponse> responseObserver) {
        testUtils.invokeStubUnaryCall(stub::commitStatus, request, responseObserver);
    }

    @Override
    public void chaincodeEvents(final SignedChaincodeEventsRequest request, final StreamObserver<ChaincodeEventsResponse> responseObserver) {
        testUtils.invokeStubServerStreamingCall(stub::chaincodeEvents, request, responseObserver);
    }

}
