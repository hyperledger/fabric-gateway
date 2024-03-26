/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.stub.ServerCallStreamObserver;
import io.grpc.stub.StreamObserver;
import org.hyperledger.fabric.protos.common.Envelope;
import org.hyperledger.fabric.protos.peer.DeliverGrpc;
import org.hyperledger.fabric.protos.peer.DeliverResponse;

/**
 * Mock Deliver gRPC service that acts as an adapter for a stub implementation.
 */
public final class MockDeliverService extends DeliverGrpc.DeliverImplBase {
    private static final TestUtils testUtils = TestUtils.getInstance();

    private final DeliverServiceStub stub;

    public MockDeliverService(final DeliverServiceStub stub) {
        this.stub = stub;
    }

    @Override
    public StreamObserver<Envelope> deliver(final StreamObserver<DeliverResponse> responseObserver) {
        synchronized (stub) {
            return testUtils.invokeStubDuplexCall(stub::blockEvents, (ServerCallStreamObserver<DeliverResponse>) responseObserver, 1);
        }
    }

    @Override
    public StreamObserver<Envelope> deliverFiltered(final StreamObserver<DeliverResponse> responseObserver) {
        synchronized (stub) {
            return testUtils.invokeStubDuplexCall(stub::filteredBlockEvents, (ServerCallStreamObserver<DeliverResponse>) responseObserver, 1);
        }
    }

    @Override
    public StreamObserver<Envelope> deliverWithPrivateData(final StreamObserver<DeliverResponse> responseObserver) {
        synchronized (stub) {
            return testUtils.invokeStubDuplexCall(stub::blockAndPrivateDataEvents, (ServerCallStreamObserver<DeliverResponse>) responseObserver, 1);
        }
    }
}
