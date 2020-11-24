/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.concurrent.Callable;

@FunctionalInterface
public interface Commit extends Callable<byte[]> {
    @Override
    byte[] call() throws ContractException;
}
