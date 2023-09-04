/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { KeyObject, X509Certificate, createHash } from 'node:crypto';
import { assertDefined } from './utils';

export function getSKIFromCertificate(certificate: X509Certificate): Buffer {
    const uncompressedPoint = getUncompressedPointOnCurve(certificate.publicKey);
    return createHash('sha256').update(uncompressedPoint).digest();
}

function getUncompressedPointOnCurve(key: KeyObject): Buffer {
    const jwk = key.export({ format: 'jwk' });
    const x = Buffer.from(assertDefined(jwk.x, 'x'), 'base64url');
    const y = Buffer.from(assertDefined(jwk.y, 'y'), 'base64url');
    const prefix = Buffer.from('04', 'hex');
    return Buffer.concat([prefix, x, y]);
}
