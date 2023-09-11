/*
 * Copyright 2023 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import BN from 'bn.js';
import { KeyObject } from 'crypto';
import { ec as EC } from 'elliptic';
import { ecSignatureAsDER } from './asn1';
import { Signer } from './signer';

const namedCurves: Record<string, EC> = {
    'P-256': new EC('p256'),
    'P-384': new EC('p384'),
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
        const signature = curve.sign(digest, privateKey, { canonical: true });
        const signatureBytes = new Uint8Array(signature.toDER());
        return Promise.resolve(signatureBytes);
    };
}

function getCurve(name: string): EC {
    const curve = namedCurves[name];
    if (!curve) {
        throw new Error(`Unsupported curve: ${name}`);
    }

    return curve;
}

export class ECSignature {
    readonly #curve: EC;
    readonly #r: BN;
    #s: BN;

    constructor(curveName: string, compactSignature: Uint8Array) {
        this.#curve = getCurve(curveName);

        const sIndex = compactSignature.length / 2;
        const r = compactSignature.slice(0, sIndex);
        const s = compactSignature.slice(sIndex);
        this.#r = new BN(r);
        this.#s = new BN(s);
    }

    normalise(): this {
        const n = this.#curve.n!;
        const halfOrder = n.divn(2);

        if (this.#s.gt(halfOrder)) {
            this.#s = n.sub(this.#s);
        }

        return this;
    }

    toDER(): Uint8Array {
        return ecSignatureAsDER(this.#r, this.#s);
    }
}
