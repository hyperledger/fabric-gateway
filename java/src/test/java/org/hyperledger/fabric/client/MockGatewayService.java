/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.stub.StreamObserver;
import org.hyperledger.fabric.protos.gateway.Event;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.ProposedTransaction;
import org.hyperledger.fabric.protos.gateway.Result;

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
    public void endorse(ProposedTransaction request, StreamObserver<PreparedTransaction> responseObserver) {
        try {
            PreparedTransaction result = stub.endorse(request);
            responseObserver.onNext(result);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public void submit(PreparedTransaction request, StreamObserver<Event> responseObserver) {
        try {
            stub.submit(request).forEach(responseObserver::onNext);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    @Override
    public void evaluate(ProposedTransaction request, StreamObserver<Result> responseObserver) {
        try {
            Result result = stub.evaluate(request);
            responseObserver.onNext(result);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }
}
