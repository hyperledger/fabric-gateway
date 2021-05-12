/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;

/**
 * Hash function implementations used to generate a digest of a supplied message.
 */
public final class Hash {
    private Hash() { }

    /**
     * SHA-256 hash the supplied message to create a digest for signing.
     * @param message Message to be hashed.
     * @return Message digest.
     */
    public static byte[] sha256(final byte[] message) {
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            return digest.digest(message);
        } catch (NoSuchAlgorithmException e) {
            // Should never happen with standard algorithm
            throw new RuntimeException(e);
        }
    }
}
