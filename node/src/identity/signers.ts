/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import crypto from 'crypto';
import { ec as EC } from 'elliptic';
import { ecPrivateKeyAsRaw } from './decoder';
import { Signer } from './signer';

const namedCurves: { [oid: string]: EC } = {
    '1.2.840.10045.3.1.7': new EC('p256'),
    '1.3.132.0.34': new EC('p384'),
};

/**
 * Create a new signing implementation that uses the supplied private key to sign messages.
 * 
 * Currently supported private key types are:
 * - NIST P-256 elliptic curve.
 * - NIST P-384 elliptic curve.
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
    const { privateKey: rawKey, curveObjectId } = ecPrivateKeyAsRaw(key);
    const curve = getCurve(curveObjectId);
    const keyPair = curve.keyFromPrivate(rawKey, 'hex');

    return async (digest) => {
        const signature = curve.sign(digest, keyPair, { canonical: true });
        const signatureBytes = new Uint8Array(signature.toDER());
        return Promise.resolve(signatureBytes);
    }
}

function getCurve(objectIdBytes: number[]): EC {
    const objectId = objectIdBytes.join('.');
    const curve = namedCurves[objectId];
    if (!curve) {
        throw new Error(`Unsupported curve object identifier: ${objectId}`);
    }

    return curve;
}
