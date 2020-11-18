/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import java.security.interfaces.ECPrivateKey;

/**
 * Utility methods for creating and manipulating identity information.
 */
public final class Signers {
    /**
     * Create a new signer that uses the supplied private key for signing.
     * @param privateKey A private key.
     * @return A signer implementation.
     */
    public static Signer newPrivateKeySigner(final ECPrivateKey privateKey) {
        return new ECDSAPrivateKeySigner(privateKey);
    }

    // Private constructor to prevent instantiation
    private Signers() { }
}
