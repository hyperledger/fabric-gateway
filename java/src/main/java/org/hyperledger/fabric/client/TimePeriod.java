/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.concurrent.TimeUnit;

/**
 * Encapsulates the value and time unit for timeouts typically required on blocking method calls in the Java
 * concurrency libraries.
 */
final class TimePeriod {
    private final long time;
    private final TimeUnit timeUnit;

    TimePeriod(final long time, final TimeUnit timeUnit) {
        this.time = time;
        this.timeUnit = timeUnit;
    }

    public long getTime() {
        return time;
    }

    public TimeUnit getTimeUnit() {
        return timeUnit;
    }

    @Override
    public String toString() {
        return String.format("%s %s", time, timeUnit);
    }
}
