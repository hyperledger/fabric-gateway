/*
 * Copyright 2023 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CurveFn } from '@noble/curves/abstract/weierstrass';
import { p256 } from '@noble/curves/p256';
import { p384 } from '@noble/curves/p384';
import { KeyObject } from 'node:crypto';
import { Signer } from './signer';

const namedCurves: Record<string, CurveFn> = {
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
        const signature = curve.sign(digest, privateKey, { lowS: true });
        return Promise.resolve(signature.toDERRawBytes());
    };
}

function getCurve(name: string): CurveFn {
    const curve = namedCurves[name];
    if (!curve) {
        throw new Error(`Unsupported curve: ${name}`);
    }

    return curve;
}
