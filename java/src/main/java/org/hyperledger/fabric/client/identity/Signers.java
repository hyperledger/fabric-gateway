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
    /**
     * Create a new signer that uses the supplied private key for signing. The {@link Identities} class provides static
     * methods to create a {@code PrivateKey} object from PEM-format data.
     * @param privateKey A private key.
     * @return A signer implementation.
     */
    public static Signer newPrivateKeySigner(final PrivateKey privateKey) {
        if (privateKey instanceof ECPrivateKey) {
            return new ECPrivateKeySigner((ECPrivateKey) privateKey);
        } else {
            throw new IllegalArgumentException("Unsupported private key type: " + privateKey.getClass().getTypeName());
        }
    }

    // Private constructor to prevent instantiation
    private Signers() { }
}
