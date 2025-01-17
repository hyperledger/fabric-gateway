/*
 * Copyright 2023 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CurveFn } from '@noble/curves/abstract/weierstrass';
import { p256 } from '@noble/curves/p256';
import { p384 } from '@noble/curves/p384';
import { Signer } from './signer';

const namedCurves: Record<string, CurveFn> = {
    'P-256': p256,
    'P-384': p384,
};

export async function newECPrivateKeySigner(key: CryptoKey): Promise<Signer> {
    if (!key.extractable) {
        throw new Error('Key is not extractable');
    }
    const { crv, d } = await crypto.subtle.exportKey('jwk', key);
    if (!crv) {
        throw new Error('Missing EC curve name');
    }
    if (!d) {
        throw new Error('Missing EC private key value');
    }

    const base64d = d.replace(/-/g, '+').replace(/_/g, '/');
    const curve = getCurve(crv);

    const keyStr = window.atob(base64d);
    const buf = new ArrayBuffer(keyStr.length);
    const bufView = new Uint8Array(buf);
    for (let i = 0, strLen = keyStr.length; i < strLen; i++) {
        bufView[i] = keyStr.charCodeAt(i);
    }
    const privateKey = new Uint8Array(buf);

    return (digest: Uint8Array) => {
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
