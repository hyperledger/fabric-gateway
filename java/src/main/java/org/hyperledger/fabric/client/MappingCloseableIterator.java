/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.Function;

final class MappingCloseableIterator<T, R> implements CloseableIterator<R> {
    private final CloseableIterator<T> iterator;
    private final Function<T, R> mapper;

    MappingCloseableIterator(final CloseableIterator<T> iterator, final Function<T, R> mapper) {
        this.iterator = iterator;
        this.mapper = mapper;
    }

    @Override
    public void close() {
        iterator.close();
    }

    @Override
    public boolean hasNext() {
        return iterator.hasNext();
    }

    @Override
    public R next() {
        T value = iterator.next();
        return mapper.apply(value);
    }
}
