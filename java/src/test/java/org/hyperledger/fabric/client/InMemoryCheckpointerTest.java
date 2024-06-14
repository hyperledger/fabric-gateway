/*
 * Copyright 2024 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

public class InMemoryCheckpointerTest extends CommonCheckpointerTest {

    @Override
    protected InMemoryCheckpointer getCheckpointerInstance() {
        return new InMemoryCheckpointer();
    }

    @Override
    protected void tearDown() {}
}
