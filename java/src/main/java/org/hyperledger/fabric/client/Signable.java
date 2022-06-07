/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * A Fabric Gateway call that can be explicitly signed. Supports off-line signing flow.
 */
public interface Signable {
    /**
     * Get the serialized message bytes.
     * Serialized bytes can be used to recreate the object using methods on {@link Gateway}.
     * @return A serialized signable object.
     */
    byte[] getBytes();

    /**
     * Get the digest of the signable object. This is used to generate a digital signature.
     * @return A hash of the signable object.
     */
    byte[] getDigest();
}
