/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable */
// @ts-nocheck

import asn1 from 'asn1.js';
import crypto from 'crypto';
import { BNInput } from 'elliptic';

const ECPrivateKey = asn1.define('ECPrivateKey', function() {
    this.seq().obj(
        this.key('version').int().def(1),
        this.key('privateKey').octstr(),
        this.key('parameters').explicit(0).objid().optional(),
        this.key('publicKey').explicit(1).bitstr().optional()
    );
});

const ECSignature = asn1.define('ECSignature', function() {
    return this.seq().obj(
        this.key('r').int(),
        this.key('s').int()
    );
});

export function ecPrivateKeyAsRaw(privateKey: crypto.KeyObject): { privateKey: Buffer, curveObjectId: number[] } {
    const privateKeyPem = privateKey.export({ format: 'der', type: 'sec1' });
    const decodedDer = ECPrivateKey.decode(privateKeyPem, 'der');
    return {
        privateKey: decodedDer.privateKey,
        curveObjectId: decodedDer.parameters,
    };
}

export function ecRawSignatureAsDer(r: BNInput, s: BNInput): Buffer {
    return ECSignature.encode({ r, s }, 'der');
}
