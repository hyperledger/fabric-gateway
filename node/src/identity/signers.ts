/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as crypto from 'crypto';
import { ec as EC } from 'elliptic';
import { ecPrivateKeyAsRaw } from './decoder';
import { Signer } from './signer';

const p256Curve = new EC('p256');

/**
 * Create a new signing implementation that uses the supplied private key to sign messages. Currently supported
 * private key types are:
 * - NIST P-256 elliptic curve.
 * @param key - A private key.
 * @returns A signing implementation.
 */
export function newPrivateKeySigner(key: crypto.KeyObject): Signer {
    if (key.type !== 'private') {
        throw new Error(`Invalid key type: ${key.type}`);
    }

    switch (key.asymmetricKeyType) {
    case 'ec':
        return newECPrivateKeySigner(key);
    default:
        throw new Error(`Unsupported private key type: ${key.asymmetricKeyType ?? 'undefined'}`);
    }
}

function newECPrivateKeySigner(key: crypto.KeyObject): Signer {
    const rawKey = ecPrivateKeyAsRaw(key);
    const keyPair = p256Curve.keyFromPrivate(rawKey, 'hex');

    return async (digest) => {
        const signature = p256Curve.sign(digest, keyPair, { canonical: true });
        const signatureBytes = new Uint8Array(signature.toDER());
        return Promise.resolve(signatureBytes);
    }
}
