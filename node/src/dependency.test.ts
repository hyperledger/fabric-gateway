/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { generateKeyPairSync } from 'node:crypto';
import { dirname } from 'node:path';
import type { signers as SignersType } from '.';

function isLoaded(moduleName: string): boolean {
    const moduleFile = require.resolve(moduleName);
    const moduleDir = dirname(moduleFile);
    return !!Object.values(require.cache).find(m => m?.filename.startsWith(moduleDir));
}

describe('optional pkcs11js dependency', () => {
    it('not loaded when accessing private key signer', () => {
        jest.resetModules();
        expect(isLoaded('pkcs11js')).toBe(false);

        const { privateKey } = generateKeyPairSync('ec', { namedCurve: 'P-256' });
        // eslint-disable-next-line @typescript-eslint/no-var-requires
        const { signers } = require('.') as { signers: typeof SignersType };
        signers.newPrivateKeySigner(privateKey);

        expect(isLoaded('pkcs11js')).toBe(false);
    });
});

// Load of @noble/hashes fails on big-endian systems
describe('@noble/hashes transitive dependency', () => {
    it('not loaded when accessing non-ECDSA private key signer', () => {
        jest.resetModules();
        expect(isLoaded('@noble/hashes')).toBe(false);

        const { privateKey } = generateKeyPairSync('ed25519');
        // eslint-disable-next-line @typescript-eslint/no-var-requires
        const { signers } = require('.') as { signers: typeof SignersType };
        signers.newPrivateKeySigner(privateKey);

        expect(isLoaded('@noble/hashes')).toBe(false);
    });
});
