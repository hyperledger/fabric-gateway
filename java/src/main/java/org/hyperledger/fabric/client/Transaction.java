/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

public interface Transaction {
    byte[] getResult();
    byte[] getBytes();
    byte[] getHash();
    Commit submitAsync();
    byte[] submitSync() throws ContractException;
}
