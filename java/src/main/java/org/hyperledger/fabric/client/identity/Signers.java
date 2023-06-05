/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import java.security.PrivateKey;
import java.security.interfaces.ECPrivateKey;

/**
 * Factory methods to create standard signing implementations.
 */
public final class Signers {
    private static final String ED25519_ALGORITHM = "Ed25519";

    /**
     * Create a new signer that uses the supplied private key for signing. The {@link Identities} class provides static
     * methods to create a {@code PrivateKey} object from PEM-format data.
     *
     * <p>Currently supported private key types are:</p>
     * <ul>
     *     <li>ECDSA.</li>
     *     <li>Ed25519.</li>
     * </ul>
     *
     * <p>Note that the Sign implementations have different expectations on the input data supplied to them.</p>
     *
     * <p>The ECDSA signers operate on a pre-computed message digest, and should be combined with an appropriate hash
     * algorithm. P-256 is typically used with a SHA-256 hash, and P-384 is typically used with a SHA-384 hash.</p>
     *
     * <p>The Ed25519 signer operates on the full message content, and should be combined with a
     * {@link org.hyperledger.fabric.client.Hash#NONE NONE} (or no-op) hash implementation to ensure the complete
     * message is passed to the signer.</p>
     * @param privateKey A private key.
     * @return A signer implementation.
     */
    public static Signer newPrivateKeySigner(final PrivateKey privateKey) {
        if (privateKey instanceof ECPrivateKey) {
            return new ECPrivateKeySigner((ECPrivateKey) privateKey);
        }

        if (ED25519_ALGORITHM.equals(privateKey.getAlgorithm())) {
            return new PrivateKeySigner(privateKey, ED25519_ALGORITHM);
        }

        throw new IllegalArgumentException("Unsupported private key type: " + privateKey.getClass().getTypeName()
                + " (" + privateKey.getAlgorithm() + ")");
    }

    // Private constructor to prevent instantiation
    private Signers() { }
}
