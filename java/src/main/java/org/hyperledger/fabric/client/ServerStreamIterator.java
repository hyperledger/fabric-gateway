/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;
import java.util.function.Supplier;

import io.grpc.Context;

final class ServerStreamIterator<T> implements CloseableIterator<T> {
    private final Context.CancellableContext context;
    private final Context previousContext;
    private final Iterator<T> iter;

    ServerStreamIterator(final Context.CancellableContext context, final Supplier<Iterator<T>> iteratorSupplier) {
        this.context = context;
        previousContext = this.context.attach();

        try {
            iter = iteratorSupplier.get();
        } catch (RuntimeException e) {
            cancelContext(e);
            throw e;
        }
    }

    private void cancelContext(final Throwable t) {
        context.detachAndCancel(previousContext, t);
    }

    @Override
    public void close() {
        cancelContext(null);
    }

    @Override
    public boolean hasNext() {
        return iter.hasNext();
    }

    @Override
    public T next() {
        return iter.next();
    }
}
