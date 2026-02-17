/*
 * Copyright 2023 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { p256, p384 } from '@noble/curves/nist.js';
import { KeyObject } from 'node:crypto';
import { Signer } from './signer.js';

type Curve = typeof p256;
const namedCurves: Record<string, Curve> = {
    'P-256': p256,
    'P-384': p384,
};

export function newECPrivateKeySigner(key: KeyObject): Signer {
    const { crv, d } = key.export({ format: 'jwk' });
    if (!crv) {
        throw new Error('Missing EC curve name');
    }
    if (!d) {
        throw new Error('Missing EC private key value');
    }

    const curve = getCurve(crv);
    const privateKey = Buffer.from(d, 'base64url');

    return (digest) => {
        const signature = curve.sign(digest, privateKey, { lowS: true, prehash: false, format: 'der' });
        return Promise.resolve(signature);
    };
}

function getCurve(name: string): Curve {
    const curve = namedCurves[name];
    if (!curve) {
        throw new Error(`Unsupported curve: ${name}`);
    }

    return curve;
}
