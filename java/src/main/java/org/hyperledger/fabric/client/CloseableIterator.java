/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;

/**
 * An iterator that can be closed when the consumer does not want to read any more elements, freeing up resources that
 * may be held by the iterator.
 * <p>Note that iteration may throw {@link GatewayRuntimeException} if the gRPC connection fails.</p>
 * @param <T> The type of elements returned by this iterator.
 */
public interface CloseableIterator<T> extends Iterator<T>, AutoCloseable {
    @Override
    void close();
}
