/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { createPublicKey, verify } from 'crypto';
import { newPrivateKeySigner } from '../src/identity/signers';
import { p256 } from '@noble/curves/p256';
import { p384 } from '@noble/curves/p384';

describe('signers', () => {
    it('throws for public key', async () => {
        const { publicKey } = await crypto.subtle.generateKey({ name: 'ECDSA', namedCurve: 'P-256' }, true, ['sign']);

        await expect(newPrivateKeySigner(publicKey)).rejects.toThrow(publicKey.type);
    });

    it('throws for unsupported private key type', async () => {
        let { privateKey } = await window.crypto.subtle.generateKey(
            {
                name: 'RSA-OAEP',
                modulusLength: 2048,
                publicExponent: new Uint8Array([1, 0, 1]),
                hash: 'SHA-256',
            },
            true,
            ['encrypt', 'decrypt'],
        );

        await expect(newPrivateKeySigner(privateKey)).rejects.toThrow(privateKey.algorithm.name);
    });

    describe('EC', () => {
        it('creates valid signer for P-256 private key', async () => {
            const { publicKey, privateKey } = await crypto.subtle.generateKey(
                { name: 'ECDSA', namedCurve: 'P-256' },
                true,
                ['sign', 'verify'],
            );
            const enc = new TextEncoder();
            const message = enc.encode('conga');

            const signer = await newPrivateKeySigner(privateKey);
            const digest = new Uint8Array(await crypto.subtle.digest('SHA-256', message));
            const signature = await signer(digest);

            const rawPublicKeyAB = await window.crypto.subtle.exportKey('raw', publicKey);
            const valid = p256.verify(signature, digest, new Uint8Array(rawPublicKeyAB));

            expect(valid).toBeTruthy();
        });

        it('creates valid signer for P-384 private key', async () => {
            const { publicKey, privateKey } = await crypto.subtle.generateKey(
                { name: 'ECDSA', namedCurve: 'P-384' },
                true,
                ['sign', 'verify'],
            );
            const enc = new TextEncoder();
            const message = enc.encode('conga');

            const signer = await newPrivateKeySigner(privateKey);
            const digest = new Uint8Array(await crypto.subtle.digest('SHA-384', message));
            const signature = await signer(digest);

            const rawPublicKeyAB = await window.crypto.subtle.exportKey('raw', publicKey);
            const valid = p384.verify(signature, digest, new Uint8Array(rawPublicKeyAB));

            expect(valid).toBeTruthy();
        });

        it('throws for unsupported curve', async () => {
            const { privateKey } = await crypto.subtle.generateKey({ name: 'ECDSA', namedCurve: 'P-521' }, true, [
                'sign',
            ]);

            await expect(newPrivateKeySigner(privateKey)).rejects.toThrow('P-521');
        });
    });

    describe('Ed25519', () => {
        it('creates valid signer', async () => {
            const { publicKey, privateKey } = await crypto.subtle.generateKey('Ed25519', false, ['sign', 'verify']);
            const message = Buffer.from('conga');

            const signer = await newPrivateKeySigner(privateKey);
            const signature = await signer(message);
            const valid = crypto.subtle.verify('Ed25519', publicKey, signature, message);

            expect(valid).toBeTruthy();
        });
    });
});
