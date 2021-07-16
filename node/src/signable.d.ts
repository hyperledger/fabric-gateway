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
     * Get the serialized commit status request message.
     */
     getBytes(): Uint8Array;

     /**
      * Get the digest of the commit status request. This is used to generate a digital signature.
      */
     getDigest(): Uint8Array;
 }
