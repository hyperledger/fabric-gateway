/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Map;

public interface Proposal {
    String getTransactionId();
    byte[] getBytes();
    byte[] getHash();
    Proposal addArguments(byte[]... args);
    Proposal addArguments(String... args);
    Proposal setTransient(Map<String, byte[]> transientData);
    byte[] evaluate();
    Transaction endorse();
}
