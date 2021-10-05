/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.concurrent.CompletableFuture;

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

public class MockGatewayService extends GatewayGrpc.GatewayImplBase {
    private static final GatewayServiceStub DEFAULT_STUB = new GatewayServiceStub();

    private final GatewayServiceStub stub;

    public MockGatewayService() {
        this(DEFAULT_STUB);
    }

    public MockGatewayService(final GatewayServiceStub stub) {
        this.stub = stub;
    }

    @Override
    public void endorse(final EndorseRequest request, final StreamObserver<EndorseResponse> responseObserver) {
        try {
            EndorseResponse response = stub.endorse(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public void submit(final SubmitRequest request, final StreamObserver<SubmitResponse> responseObserver) {
        try {
            SubmitResponse response = stub.submit(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public void evaluate(final EvaluateRequest request, final StreamObserver<EvaluateResponse> responseObserver) {
        try {
            EvaluateResponse response = stub.evaluate(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public void commitStatus(final SignedCommitStatusRequest request, final StreamObserver<CommitStatusResponse> responseObserver) {
        try {
            CommitStatusResponse response = stub.commitStatus(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public void chaincodeEvents(final SignedChaincodeEventsRequest request, final StreamObserver<ChaincodeEventsResponse> responseObserver) {
        CompletableFuture.runAsync(() -> {
            try {
                stub.chaincodeEvents(request).forEachOrdered(responseObserver::onNext);
                responseObserver.onCompleted();
            } catch (Throwable t) {
                responseObserver.onError(t);
            }
        });
    }
}
