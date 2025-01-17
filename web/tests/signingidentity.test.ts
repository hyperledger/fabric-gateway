/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { msp } from '@hyperledger/fabric-protos';
import { Identity, Signer } from '../src';
import { SigningIdentity } from '../src/signingidentity';

const utf8Encoder = new TextEncoder();
const utf8Decoder = new TextDecoder();

describe('SigningIdentity', () => {
    const signature = 'SIGNATURE';

    let identity: Identity;
    let signer: Signer;

    beforeEach(() => {
        identity = {
            mspId: 'MSP_ID',
            credentials: utf8Encoder.encode('CREDENTIALS'),
        };
        signer = () => Promise.resolve(utf8Encoder.encode(signature));
    });

    describe('identity', () => {
        it('changes to returned identity do not modify signing identity', () => {
            const expectedMspId = identity.mspId;
            const expectedCredentials = Uint8Array.from(identity.credentials); // Copy
            const signingIdentity = new SigningIdentity({ identity, signer });

            const output = signingIdentity.getIdentity();
            output.mspId = 'wrong';
            output.credentials.fill(0);

            const actual = signingIdentity.getIdentity();
            expect(actual.mspId).toBe(expectedMspId);
            const actualCredentials = Uint8Array.from(actual.credentials); // Ensure it's really a Uint8Array
            expect(actualCredentials).toEqual(expectedCredentials);
        });

        it('changes to supplied identity do not modify signing identity', () => {
            const expectedMspId = identity.mspId;
            const expectedCredentials = Uint8Array.from(identity.credentials); // Copy

            const signingIdentity = new SigningIdentity({ identity, signer });
            identity.mspId = 'wrong';
            identity.credentials.fill(0);

            const actual = signingIdentity.getIdentity();
            expect(actual.mspId).toBe(expectedMspId);
            const actualCredentials = Uint8Array.from(actual.credentials); // Ensure it's really a Uint8Array
            expect(actualCredentials).toEqual(expectedCredentials);
        });
    });

    describe('creator', () => {
        it('returns a valid SerializedIdentity protobuf', () => {
            const signingIdentity = new SigningIdentity({ identity, signer });

            const creator = signingIdentity.getCreator();

            const actual = msp.SerializedIdentity.deserializeBinary(creator);
            expect(actual.getMspid()).toBe(identity.mspId);
            const credentials = Uint8Array.from(actual.getIdBytes_asU8()); // Ensure it's really a Uint8Array
            expect(credentials).toEqual(identity.credentials);
        });

        it('changes to returned creator do not modify signing identity', () => {
            const signingIdentity = new SigningIdentity({ identity, signer });
            const expected = Uint8Array.from(signingIdentity.getCreator()); // Ensure it's really a Uint8Array

            const creator = signingIdentity.getCreator();
            creator.fill(0);

            const actual = Uint8Array.from(signingIdentity.getCreator()); // Ensure it's really a Uint8Array
            expect(actual).toEqual(expected);
        });
    });

    describe('signing', () => {
        it('uses supplied signer', async () => {
            const message = utf8Encoder.encode('MESSAGE');
            const signingIdentity = new SigningIdentity({ identity, signer });

            const result = await signingIdentity.sign(message);

            const actual = utf8Decoder.decode(result);
            expect(actual).toEqual(signature);
        });
    });
});
