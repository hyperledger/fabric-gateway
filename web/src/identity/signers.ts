/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { newECPrivateKeySigner } from './ecdsa';
import { Signer } from './signer';

/**
 * Create a new signing implementation that uses the supplied private key to sign messages.
 *
 * Currently supported private key types are:
 * - NIST P-256 elliptic curve.
 * - NIST P-384 elliptic curve.
 * - Ed25519.
 *
 * Note that the signer implementations have different expectations on the input data supplied to them.
 *
 * The P-256 and P-384 signers operate on a pre-computed message digest, and should be combined with an appropriate
 * hash algorithm. P-256 is typically used with a SHA-256 hash, and P-384 is typically used with a SHA-384 hash.
 *
 * The Ed25519 signer operates on the full message content, and should be combined with a {@link hash.none | none} (or
 * no-op) hash implementation to ensure the complete message is passed to the signer.
 *
 * @param key - A private key.
 * @returns A signing implementation.
 */
export async function newPrivateKeySigner(key: CryptoKey): Promise<Signer> {
    if (key.type !== 'private') {
        throw new Error(`Invalid key type: ${key.type}`);
    }

    switch (key.algorithm.name) {
        case 'ECDSA':
            return newECPrivateKeySigner(key);
        case 'Ed25519':
            return async (message: Uint8Array) => {
                const signature = new Uint8Array(await crypto.subtle.sign('Ed25519', key, message));
                return Promise.resolve(signature);
            };
        default:
            throw new Error(`Unsupported private key type: ${String(key.algorithm.name)}`);
    }
}
