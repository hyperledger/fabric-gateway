/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Signer } from './signer'
import { ec as EC } from 'elliptic';
import { KEYUTIL } from 'jsrsasign';
import * as crypto from 'crypto';


const p256Curve = new EC('p256');

export function newECDSAPrivateKeySigner(privateKey: crypto.KeyObject): Signer {
    const privateKeyPem = privateKey.export({ format: 'pem', type: 'pkcs8' });
    const key = KEYUTIL.getKey(privateKeyPem.toString()) as any; // eslint-disable-line @typescript-eslint/no-explicit-any
    const keyPair = p256Curve.keyFromPrivate(key.prvKeyHex, 'hex'); // TODO: key.prvKeyHex is an undocumented internal

    return async (digest) => {
        const signature = p256Curve.sign(digest, keyPair, { canonical: true });
        return signature.toDER();
    }
}
