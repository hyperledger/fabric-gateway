/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { createHash, generateKeyPairSync, verify } from 'crypto';
import { newPrivateKeySigner } from './signers';

describe('signers', () => {
    it('throws for public key', () => {
        const { publicKey } = generateKeyPairSync('ec', { namedCurve: 'P-256' });
        expect(() => newPrivateKeySigner(publicKey))
            .toThrowError(publicKey.type);
    });

    it('throws for unsupported private key type', () => {
        const { privateKey } = generateKeyPairSync('dsa', { modulusLength: 2048, divisorLength: 256 });

        expect(() => newPrivateKeySigner(privateKey))
            .toThrowError(privateKey.asymmetricKeyType);
    });

    describe('EC', () => {
        it('creates valid signer for P-256 private key', async () => {
            const { publicKey, privateKey } = generateKeyPairSync('ec', { namedCurve: 'P-256' });
            const message = Buffer.from('conga');

            const signer = newPrivateKeySigner(privateKey);
            const digest = createHash('sha256').update(message).digest();
            const signature = await signer(digest);
            const valid = verify('sha256', message, publicKey, signature);

            expect(valid).toBeTruthy();
        });

        it('creates valid signer for P-384 private key', async () => {
            const { publicKey, privateKey } = generateKeyPairSync('ec', { namedCurve: 'P-384' });
            const message = Buffer.from('conga');

            const signer = newPrivateKeySigner(privateKey);
            const digest = createHash('sha256').update(message).digest();
            const signature = await signer(digest);
            const valid = verify('sha256', message, publicKey, signature);

            expect(valid).toBeTruthy();
        });

        it('throws for unsupported curve', () => {
            const { privateKey } = generateKeyPairSync('ec', { namedCurve: 'secp256k1' });
            expect(() => newPrivateKeySigner(privateKey))
                .toThrowError('1.3.132.0.10');
        });
    });
});
