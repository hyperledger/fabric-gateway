/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.stub.StreamObserver;
import org.hyperledger.fabric.gateway.Event;
import org.hyperledger.fabric.gateway.GatewayGrpc;
import org.hyperledger.fabric.gateway.PreparedTransaction;
import org.hyperledger.fabric.gateway.ProposedTransaction;
import org.hyperledger.fabric.gateway.Result;

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
        PreparedTransaction result = stub.endorse(request);
        responseObserver.onNext(result);
        responseObserver.onCompleted();
    }

    @Override
    public void submit(PreparedTransaction request, StreamObserver<Event> responseObserver) {
        stub.submit(request).forEach(responseObserver::onNext);
        responseObserver.onCompleted();
    }

    @Override
    public void evaluate(ProposedTransaction request, StreamObserver<Result> responseObserver) {
        Result result = stub.evaluate(request);
        responseObserver.onNext(result);
        responseObserver.onCompleted();
    }
}
