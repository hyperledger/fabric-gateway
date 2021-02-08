/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as crypto from 'crypto';
import { newPrivateKeySigner } from './signers';

describe('signers', () => {
    const { publicKey, privateKey } = crypto.generateKeyPairSync('ec', { namedCurve: 'P-256' });

    it('throws for public key', () => {
        expect(() => newPrivateKeySigner(publicKey))
            .toThrowError(publicKey.type);
    });

    it('throws for unsupported private key type', () => {
        const { privateKey: dsaKey } = crypto.generateKeyPairSync('dsa', { modulusLength: 2048, divisorLength: 256 });

        expect(() => newPrivateKeySigner(dsaKey))
            .toThrowError(dsaKey.asymmetricKeyType);
    });

    it('creates valid signer for EC private key', async () => {
        const message = Buffer.from('conga');

        const signer = newPrivateKeySigner(privateKey);
        const digest = crypto.createHash('sha256').update(message).digest();
        const signature = await signer(digest);
        const valid = crypto.verify('sha256', message, publicKey, signature);

        expect(valid).toBeTruthy();
    });
});