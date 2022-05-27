/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.UnaryOperator;

import io.grpc.CallOptions;

/**
 * A Fabric Gateway call to obtain events.
 * @param <T> The type of events obtained by this request.
 */
public interface EventsRequest<T> extends Signable {
    /**
     * Get events. The Java gRPC implementation may not begin reading events until the first use of the returned
     * iterator.
     * <p>Note that the returned iterator may throw {@link GatewayRuntimeException} during iteration if a gRPC
     * connection error occurs.</p>
     * @return Ordered sequence of events.
     */
    default CloseableIterator<T> getEvents() {
        return getEvents(GatewayUtils.asCallOptions());
    }

    /**
     * Get events. The Java gRPC implementation may not begin reading events until the first use of the returned
     * iterator.
     * <p>Note that the returned iterator may throw {@link GatewayRuntimeException} during iteration if a gRPC
     * connection error occurs.</p>
     * @param options Function that transforms call options.
     * @return Ordered sequence of events.
     */
    CloseableIterator<T> getEvents(UnaryOperator<CallOptions> options);

    /**
     * Get events. The Java gRPC implementation may not begin reading events until the first use of the returned
     * iterator.
     * <p>Note that the returned iterator may throw {@link GatewayRuntimeException} during iteration if a gRPC
     * connection error occurs.</p>
     * @param options Call options.
     * @return Ordered sequence of events.
     * @deprecated Replaced by {@link #getEvents(UnaryOperator)}.
     */
    @Deprecated
    default CloseableIterator<T> getEvents(CallOption... options) {
        return getEvents(GatewayUtils.asCallOptions(options));
    }
}
