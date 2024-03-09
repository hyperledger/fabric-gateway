/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { msp } from '@hyperledger/fabric-protos';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';
import { SigningIdentity } from './signingidentity';

describe('SigningIdentity', () => {
    let identity: Identity;

    beforeEach(() => {
        identity = {
            mspId: 'MSP_ID',
            credentials: new Uint8Array(Buffer.from('CREDENTIALS')),
        };
    });

    describe('identity', () => {
        it('changes to returned identity do not modify signing identity', () => {
            const expectedMspId = identity.mspId;
            const expectedCredentials = new Uint8Array(identity.credentials); // Copy
            const signingIdentity = new SigningIdentity({ identity });

            const output = signingIdentity.getIdentity();
            output.mspId = 'wrong';
            output.credentials.fill(0);

            const actual = signingIdentity.getIdentity();
            expect(actual.mspId).toBe(expectedMspId);
            const actualCredentials = new Uint8Array(actual.credentials); // Ensure it's really a Uint8Array
            expect(actualCredentials).toEqual(expectedCredentials);
        });

        it('changes to supplied identity do not modify signing identity', () => {
            const expectedMspId = identity.mspId;
            const expectedCredentials = new Uint8Array(identity.credentials); // Copy

            const signingIdentity = new SigningIdentity({ identity });
            identity.mspId = 'wrong';
            identity.credentials.fill(0);

            const actual = signingIdentity.getIdentity();
            expect(actual.mspId).toBe(expectedMspId);
            const actualCredentials = new Uint8Array(actual.credentials); // Ensure it's really a Uint8Array
            expect(actualCredentials).toEqual(expectedCredentials);
        });
    });

    describe('creator', () => {
        it('returns a valid SerializedIdentity protobuf', () => {
            const signingIdentity = new SigningIdentity({ identity });

            const creator = signingIdentity.getCreator();

            const actual = msp.SerializedIdentity.deserializeBinary(creator);
            expect(actual.getMspid()).toBe(identity.mspId);
            const credentials = new Uint8Array(actual.getIdBytes_asU8()); // Ensure it's really a Uint8Array
            expect(credentials).toEqual(identity.credentials);
        });

        it('changes to returned creator do not modify signing identity', () => {
            const signingIdentity = new SigningIdentity({ identity });
            const expected = new Uint8Array(signingIdentity.getCreator()); // Ensure it's really a Uint8Array

            const creator = signingIdentity.getCreator();
            creator.fill(0);

            const actual = new Uint8Array(signingIdentity.getCreator()); // Ensure it's really a Uint8Array
            expect(actual).toEqual(expected);
        });
    });

    describe('signing', () => {
        it('default signer throws', async () => {
            const signingIdentity = new SigningIdentity({ identity });
            const digest = Buffer.from('DIGEST');

            await expect(signingIdentity.sign(digest)).rejects.toThrow();
        });

        it('uses supplied signer', async () => {
            const expected = new Uint8Array(Buffer.from('SIGNATURE'));
            const signer: Signer = async () => Promise.resolve(expected);
            const digest = Buffer.from('DIGEST');
            const signingIdentity = new SigningIdentity({ identity, signer });

            const actual = await signingIdentity.sign(digest);

            expect(actual).toEqual(expected);
        });
    });

    describe('hashing', () => {
        it('hashes of identical values are identical', () => {
            const message = Buffer.from('MESSAGE');
            const signingIdentity = new SigningIdentity({ identity });

            const first = signingIdentity.hash(message);
            const second = signingIdentity.hash(message);

            expect(first).toEqual(second);
        });

        it('hashes of different values are different', () => {
            const message1 = Buffer.from('FOO');
            const message2 = Buffer.from('BAR');
            const signingIdentity = new SigningIdentity({ identity });

            const first = signingIdentity.hash(message1);
            const second = signingIdentity.hash(message2);

            expect(first).not.toEqual(second);
        });
    });
});
