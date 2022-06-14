/*
 *  Copyright 2022 IBM All Rights Reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.stream.Stream;

import org.hyperledger.fabric.protos.common.Envelope;
import org.hyperledger.fabric.protos.peer.DeliverResponse;

/**
 * Simplified stub implementation for Deliver gRPC service, to be used as a spy by unit tests.
 */
public class DeliverServiceStub {
    public Stream<DeliverResponse> blockEvents(final Stream<Envelope> requests) {
        return Stream.empty();
    }

    public Stream<DeliverResponse> filteredBlockEvents(final Stream<Envelope> requests) {
        return Stream.empty();
    }

    public Stream<DeliverResponse> blockAndPrivateDataEvents(final Stream<Envelope> requests) {
        return Stream.empty();
    }
}
