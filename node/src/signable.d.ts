/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * A call that can be explicitly signed. Supports off-line signing flow.
 */
export interface Signable {
    /**
     * Get the serialized bytes of the signable object.
     * Serialized bytes can be used to recreate the object using methods on {@link Gateway}.
     */
    getBytes(): Uint8Array;

    /**
     * Get the digest of the signable object. This is used to generate a digital signature.
     */
    getDigest(): Uint8Array;
}
