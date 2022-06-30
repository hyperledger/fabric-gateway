/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Objects;
import java.util.concurrent.TimeUnit;
import java.util.function.UnaryOperator;

import io.grpc.CallOptions;
import io.grpc.Deadline;

/**
 * Options defining runtime behavior of a gRPC service invocation.
 * @deprecated Use gRPC {@link io.grpc.CallOptions} instead.
 */
@Deprecated
public final class CallOption {
    private final UnaryOperator<CallOptions> operator;

    private CallOption(final UnaryOperator<CallOptions> operator) {
        this.operator = operator;
    }

    /**
     * An absolute deadline.
     * @param deadline the deadline.
     * @return a call option.
     */
    public static CallOption deadline(final Deadline deadline) {
        Objects.requireNonNull(deadline, "deadline");
        return new CallOption(options -> options.withDeadline(deadline));
    }

    /**
     * A deadline that is after the given duration from when the call is initiated.
     * @param duration a time duration.
     * @param unit units for the time duration.
     * @return a call option.
     */
    public static CallOption deadlineAfter(final long duration, final TimeUnit unit) {
        Objects.requireNonNull(unit, "unit");
        return new CallOption(options -> options.withDeadlineAfter(duration, unit));
    }

    CallOptions apply(final CallOptions options) {
        return operator.apply(options);
    }
}
