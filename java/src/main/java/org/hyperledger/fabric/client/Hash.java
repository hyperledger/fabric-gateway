/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.function.Function;

/**
 * Hash function implementations used to generate a digest of a supplied message.
 */
public enum Hash implements Function<byte[], byte[]> {
    /**
     * Returns the input message unchanged. This can be used if the signing implementation requires the full message
     * bytes, not just a pre-generated digest, such as Ed25519.
     */
    NONE(Function.identity()),

    /** SHA-256 hash. */
    SHA256(message -> digest("SHA-256", message)),

    /** SHA-384 hash. */
    SHA384(message -> digest("SHA-384", message)),

    /** SHA3-256 hash. */
    SHA3_256(message -> digest("SHA3-256", message)),

    /** SHA3-384 hash. */
    SHA3_384(message -> digest("SHA3-384", message));

    /**
     * SHA-256 hash the supplied message to create a digest for signing.
     * @deprecated Replaced by {@link #SHA256}
     * @param message Message to be hashed.
     * @return Message digest.
     */
    @Deprecated
    public static byte[] sha256(final byte[] message) {
        return SHA256.apply(message);
    }

    private final Function<byte[], byte[]> implementation;

    Hash(final Function<byte[], byte[]> implementation) {
        this.implementation = implementation;
    }

    /**
     * Hash the supplied message to create a digest for signing.
     * @param message Message to be hashed.
     * @return Message digest.
     */
    public byte[] apply(final byte[] message) {
        return implementation.apply(message);
    }

    private static byte[] digest(final String algorithm, final byte[] message) {
        try {
            MessageDigest digest = MessageDigest.getInstance(algorithm);
            return digest.digest(message);
        } catch (NoSuchAlgorithmException e) {
            // Should never happen with standard algorithm
            throw new RuntimeException(e);
        }
    }
}
