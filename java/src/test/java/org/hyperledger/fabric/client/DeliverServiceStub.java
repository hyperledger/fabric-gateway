/*
 *  Copyright 2022 IBM All Rights Reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.stream.Stream;

import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.peer.EventsPackage;

/**
 * Simplified stub implementation for Deliver gRPC service, to be used as a spy by unit tests.
 */
public class DeliverServiceStub {
    public Stream<EventsPackage.DeliverResponse> blockEvents(final Stream<Common.Envelope> requests) {
        return Stream.empty();
    }

    public Stream<EventsPackage.DeliverResponse> filteredBlockEvents(final Stream<Common.Envelope> requests) {
        return Stream.empty();
    }

    public Stream<EventsPackage.DeliverResponse> blockAndPrivateDataEvents(final Stream<Common.Envelope> requests) {
        return Stream.empty();
    }
}
