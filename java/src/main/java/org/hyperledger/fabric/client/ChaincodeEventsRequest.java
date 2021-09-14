/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;

/**
 * A Fabric Gateway call to obtain chaincode events. Supports off-line signing flow using
 * {@link Network#newSignedChaincodeEventsRequest(byte[], byte[])}.
 */
public interface ChaincodeEventsRequest extends Signable {
    /**
     * Get events emitted by transaction functions of a specific chaincode. Note that the returned {@link Iterator} may
     * throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC connection error occurs.
     * @return Ordered sequence of events.
     */
    Iterator<ChaincodeEvent> getEvents();
}
