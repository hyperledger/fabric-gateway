/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable */
// @ts-nocheck

import { define } from 'asn1.js';
import { KeyObject } from 'crypto';

const ECPrivateKey = define('ECPrivateKey', function() {
    this.seq().obj(
        this.key('version').int().def(1),
        this.key('privateKey').octstr(),
        this.key('parameters').explicit(0).objid().optional(),
        this.key('publicKey').explicit(1).bitstr().optional()
    );
});

export function ecPrivateKeyAsRaw(privateKey: KeyObject): { privateKey: Buffer, curveObjectId: number[] } {
    const privateKeyPem = privateKey.export({ format: 'der', type: 'sec1' });
    const decodedDer = ECPrivateKey.decode(privateKeyPem, 'der');
    return {
        privateKey: decodedDer.privateKey,
        curveObjectId: decodedDer.parameters,
    };
}
