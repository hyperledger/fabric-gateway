/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { KeyObject, sign } from 'node:crypto';
import { newECPrivateKeySigner } from './ecdsa';
import { HSMSignerFactory, type HSMSignerFactoryImpl as HSMSignerFactoryImplType } from './hsmsigner';
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
 * The Ed25519 signer operates on the full message content, and should be combined with a `none` (or no-op) hash
 * implementation to ensure the complete message is passed to the signer.
 *
 * @param key - A private key.
 * @returns A signing implementation.
 */
export function newPrivateKeySigner(key: KeyObject): Signer {
    if (key.type !== 'private') {
        throw new Error(`Invalid key type: ${key.type}`);
    }

    switch (key.asymmetricKeyType) {
        case 'ec':
            return newECPrivateKeySigner(key);
        case 'ed25519':
            return newNodePrivateKeySigner(key);
        default:
            throw new Error(`Unsupported private key type: ${String(key.asymmetricKeyType)}`);
    }
}

function newNodePrivateKeySigner(key: KeyObject, hashAlgorithm?: string): Signer {
    return (message) => {
        const signature = sign(hashAlgorithm, message, key);
        return Promise.resolve(signature);
    };
}

/**
 * Create an HSM Signer factory. A single signer factory instance should be used to create all required HSM signers.
 */
export function newHSMSignerFactory(library: string): HSMSignerFactory {
    if (!library || library.trim() === '') {
        throw new Error('library must be provided');
    }

    // Dynamic module load to prevent unnecessary load of optional pkcs11js dependency
    const { HSMSignerFactoryImpl } = require('./hsmsigner') as {
        HSMSignerFactoryImpl: typeof HSMSignerFactoryImplType;
    };

    return new HSMSignerFactoryImpl(library);
}
