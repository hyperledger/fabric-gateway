/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;

final class Hash {
    public static byte[] sha256(byte[] message) {
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            return digest.digest(message);
        } catch (NoSuchAlgorithmException e) {
            // Should never happen with standard algorithm
            throw new RuntimeException(e);
        }
    }
}
