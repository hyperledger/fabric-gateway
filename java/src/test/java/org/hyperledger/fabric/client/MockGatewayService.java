/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.gateway.CommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.CommitStatusResponse;
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.EndorseResponse;
import org.hyperledger.fabric.protos.gateway.EvaluateRequest;
import org.hyperledger.fabric.protos.gateway.EvaluateResponse;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.hyperledger.fabric.protos.gateway.SubmitResponse;

import io.grpc.stub.StreamObserver;

public class MockGatewayService extends GatewayGrpc.GatewayImplBase {
    private static final GatewayServiceStub DEFAULT_STUB = new GatewayServiceStub();

    private final GatewayServiceStub stub;

    public MockGatewayService() {
        this(DEFAULT_STUB);
    }

    public MockGatewayService(GatewayServiceStub stub) {
        this.stub = stub;
    }

    @Override
    public void endorse(EndorseRequest request, StreamObserver<EndorseResponse> responseObserver) {
        try {
            EndorseResponse response = stub.endorse(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public void submit(SubmitRequest request, StreamObserver<SubmitResponse> responseObserver) {
        try {
            SubmitResponse response = stub.submit(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public void evaluate(EvaluateRequest request, StreamObserver<EvaluateResponse> responseObserver) {
        try {
            EvaluateResponse response = stub.evaluate(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public void commitStatus(CommitStatusRequest request, StreamObserver<CommitStatusResponse> responseObserver) {
        try {
            CommitStatusResponse response = stub.commitStatus(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }
}
