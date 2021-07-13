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
     * Get the serialized proposal message bytes.
     * @return A serialized proposal.
     */
    byte[] getBytes();

    /**
     * Get the digest of the serialized proposal. This is used to generate a digital signature.
     * @return A hash of the proposal.
     */
    byte[] getDigest();
}
