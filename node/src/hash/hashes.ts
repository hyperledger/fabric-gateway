/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { createHash } from 'crypto';
import { Hash } from './hash';

/**
 * Returns the input message unchanged. This can be used if the signing implementation requires the full message bytes,
 * not just a pre-generated digest, such as Ed25519.
 */
export const none: Hash = (message) => message;

/**
 * SHA256 hash the supplied message bytes to create a digest for signing.
 */
export const sha256: Hash = (message) => digest('sha256', message);

/**
 * SHA384 hash the supplied message bytes to create a digest for signing.
 */
export const sha384: Hash = (message) => digest('sha384', message);

/**
 * SHA3-256 hash the supplied message bytes to create a digest for signing.
 */
export const sha3_256: Hash = (message) => digest('sha3-256', message);

/**
 * SHA3-384 hash the supplied message bytes to create a digest for signing.
 */
export const sha3_384: Hash = (message) => digest('sha3-384', message);

function digest(algorithm: string, message: Uint8Array): Uint8Array {
    return createHash(algorithm).update(message).digest();
}
