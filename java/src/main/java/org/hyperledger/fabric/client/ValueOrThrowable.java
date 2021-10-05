/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;

final class ValueOrThrowable<T> {
    private final T value;
    private final Throwable throwable;

    ValueOrThrowable(final T value) {
        Objects.requireNonNull(value);
        this.value = value;
        throwable = null;
    }

    ValueOrThrowable(final Throwable t) {
        Objects.requireNonNull(t);
        value = null;
        throwable = t;
    }

    public T getValue() {
        return value;
    }

    public Throwable getThrowable() {
        return throwable;
    }
}
