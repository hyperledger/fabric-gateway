/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;
import java.util.concurrent.TimeUnit;
import java.util.function.UnaryOperator;

import io.grpc.Deadline;
import io.grpc.stub.AbstractStub;

/**
 * Options defining runtime behavior of a gRPC service invocation.
 */
public final class CallOption {
    private final UnaryOperator<AbstractStub<?>> operator;

    private CallOption(final UnaryOperator<AbstractStub<?>> operator) {
        this.operator = operator;
    }

    /**
     * An absolute deadline.
     * @param deadline the deadline.
     * @return a call option.
     */
    public static CallOption deadline(final Deadline deadline) {
        Objects.requireNonNull(deadline, "deadline");
        return new CallOption(stub -> stub.withDeadline(deadline));
    }

    /**
     * A deadline that is after the given duration from now.
     * @param duration a time duration.
     * @param unit units for the time duration.
     * @return a call option.
     */
    public static CallOption deadlineAfter(final long duration, final TimeUnit unit) {
        Objects.requireNonNull(unit, "unit");
        return new CallOption(stub -> stub.withDeadlineAfter(duration, unit));
    }

    @SuppressWarnings("unchecked")
    <T extends AbstractStub<T>> T apply(final T stub) {
        return (T) operator.apply(stub);
    }
}
