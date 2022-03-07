/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * A builder used to create new object instances from configuration state.
 * @param <T> The type of object built.
 */
public interface Builder<T> {
    /**
     * Build an instance.
     * @return A built instance.
     */
    T build();
}
