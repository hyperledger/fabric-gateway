/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.stub.ServerCallStreamObserver;
import io.grpc.stub.StreamObserver;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.peer.DeliverGrpc;
import org.hyperledger.fabric.protos.peer.EventsPackage;

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
    public StreamObserver<Common.Envelope> deliver(final StreamObserver<EventsPackage.DeliverResponse> responseObserver) {
        return testUtils.invokeStubDuplexCall(stub::blockEvents, (ServerCallStreamObserver<EventsPackage.DeliverResponse>) responseObserver, 1);
    }

    @Override
    public StreamObserver<Common.Envelope> deliverFiltered(final StreamObserver<EventsPackage.DeliverResponse> responseObserver) {
        return testUtils.invokeStubDuplexCall(stub::filteredBlockEvents, (ServerCallStreamObserver<EventsPackage.DeliverResponse>) responseObserver, 1);
    }

    @Override
    public StreamObserver<Common.Envelope> deliverWithPrivateData(final StreamObserver<EventsPackage.DeliverResponse> responseObserver) {
        return testUtils.invokeStubDuplexCall(stub::blockAndPrivateDataEvents, (ServerCallStreamObserver<EventsPackage.DeliverResponse>) responseObserver, 1);
    }
}
