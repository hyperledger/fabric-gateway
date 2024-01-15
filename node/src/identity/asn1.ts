/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable */
// @ts-nocheck

import { define } from 'asn1.js';
import BN from 'bn.js';

const ECSignature = define('ECSignature', function () {
    return this.seq().obj(this.key('r').int(), this.key('s').int());
});

export function ecSignatureAsDER(r: BN, s: BN): Uint8Array {
    return new Uint8Array(ECSignature.encode({ r, s }, 'der'));
}
