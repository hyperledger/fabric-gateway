/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, jest } from '@jest/globals';
import { p256 } from '@noble/curves/nist';
import { createHash } from 'node:crypto';
import pkcs11js, { Handle, InitializationOptions, Mechanism, Pkcs11Error, Template, TokenInfo } from 'pkcs11js';
import { HSMSignerOptions } from './hsmsigner';
import { newHSMSignerFactory } from './signers';

const hsmOptions: HSMSignerOptions = {
    label: 'ForFabric',
    pin: '98765432',
    identifier: 'id',
};

const mocks = {
    load: jest.fn<(path: string) => void>(),
    C_Initialize: jest.fn<(options?: InitializationOptions) => void>(),
    close: jest.fn<() => void>(),
    C_Finalize: jest.fn<() => void>(),
    C_OpenSession: jest.fn<(slot: Handle, flags: number) => Handle>(),
    C_CloseSession: jest.fn<(session: Handle) => void>(),
    C_SignInit: jest.fn<(session: Handle, mechanism: Mechanism, key: Handle) => void>(),
    C_SignAsync: jest.fn<(session: Handle, inData: Buffer, outData: Buffer) => Promise<Buffer>>(),
    C_GetSlotList: jest.fn<(tokenPresent?: boolean) => Handle[]>(),
    C_GetTokenInfo: jest.fn<(slot: Handle) => TokenInfo>(),
    C_Login: jest.fn<(session: Handle, userType: number, pin?: string) => void>(),
    C_FindObjectsInit: jest.fn<(session: Handle, template: Template) => void>(),
    C_FindObjects: jest.fn<(session: Handle, maxObjectCount: number) => Handle[]>(),
    C_FindObjectsFinal: jest.fn<(session: Handle) => void>(),
};

function resetMocks(): void {
    Object.values(mocks).forEach((mock) => mock.mockReset());
}

jest.mock('pkcs11js', () => {
    const originalModule = jest.requireActual<typeof pkcs11js>('pkcs11js');

    return {
        ...originalModule,
        PKCS11: jest.fn().mockImplementation(() => mocks),
    };
});

describe('when creating or disposing of an HSM Signer Factory', () => {
    beforeEach(() => {
        resetMocks();
    });

    it('throws if library option is not valid', () => {
        mocks.C_Initialize.mockImplementation(() => {
            throw new Error('Some Error');
        });
        expect(() => newHSMSignerFactory('somelibrary')).toThrow('Some Error');
    });

    it('throws if library option is not provided', () => {
        expect(() => newHSMSignerFactory('')).toThrow('library must be provided');
    });

    it('can be disposed', () => {
        const hsmSignerFactory = newHSMSignerFactory('somelibrary');
        expect(() => {
            hsmSignerFactory.dispose();
        }).not.toThrow();
    });
});

describe('When using an HSM Signer', () => {
    const slot1 = Buffer.from('1234');
    const slot2 = Buffer.from('5678');

    const mockSession = Buffer.from('mockSession');
    const mockPrivateKeyHandle = Buffer.from('someobject');

    const hsmSignerFactory = newHSMSignerFactory('somelibrary');

    const privateKey = p256.utils.randomSecretKey();
    const publicKey = p256.getPublicKey(privateKey);

    beforeEach(() => {
        resetMocks();
        mocks.C_GetTokenInfo.mockImplementation((slot: Buffer): TokenInfo => {
            if (Buffer.compare(slot, slot1) === 0) {
                return { label: 'ForFabric' } as TokenInfo;
            }
            return { label: 'someLabel' } as TokenInfo;
        });
        mocks.C_GetSlotList.mockReturnValue([slot1, slot2]);
        mocks.C_OpenSession.mockReturnValue(mockSession);
        mocks.C_FindObjects.mockReturnValue([mockPrivateKeyHandle]);
        mocks.C_SignAsync.mockImplementation((session: Buffer, digest: Buffer, buffer: Buffer) => {
            const signature = p256.sign(digest, privateKey).toBytes('compact');
            signature.forEach((b, i) => buffer.writeUInt8(b, i));
            // Return buffer of exactly signature length regardless of supplied buffer size
            const result = buffer.subarray(0, signature.length);
            return Promise.resolve(result);
        });
    });

    it('throws if label, pin or identifier are blank or not provided', () => {
        const badHSMOptions: HSMSignerOptions = {
            label: '',
            pin: '98765432',
            identifier: 'id',
        };

        expect(() => hsmSignerFactory.newSigner(badHSMOptions)).toThrow('label property must be provided');

        badHSMOptions.label = 'ForFabric';
        badHSMOptions.pin = '';
        expect(() => hsmSignerFactory.newSigner(badHSMOptions)).toThrow('pin property must be provided');

        badHSMOptions.pin = '98765432';
        badHSMOptions.identifier = '';
        expect(() => hsmSignerFactory.newSigner(badHSMOptions)).toThrow('identifier property must be provided');

        const noLabelOptions = {
            pin: '98765432',
            identifier: 'id',
        };
        expect(() => hsmSignerFactory.newSigner(noLabelOptions as HSMSignerOptions)).toThrow(
            'label property must be provided',
        );

        const noPinOptions = {
            label: 'ForFabric',
            identifier: 'id',
        };
        expect(() => hsmSignerFactory.newSigner(noPinOptions as HSMSignerOptions)).toThrow(
            'pin property must be provided',
        );

        const noIdentifierOptions = {
            label: 'ForFabric',
            pin: '98765432',
        };
        expect(() => hsmSignerFactory.newSigner(noIdentifierOptions as HSMSignerOptions)).toThrow(
            'identifier property must be provided',
        );
    });

    it('throws an error if no slots are returned', () => {
        mocks.C_GetSlotList.mockReturnValue([]);
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).toThrow('No pkcs11 slots can be found');
    });

    it('throws an error if label cannot be found and there are slots', () => {
        const badHSMOptions: HSMSignerOptions = {
            label: 'someunknownlabel',
            pin: '98765432',
            identifier: 'id',
        };

        expect(() => hsmSignerFactory.newSigner(badHSMOptions)).toThrow(
            'label someunknownlabel cannot be found in the pkcs11 slot list',
        );
    });

    it('finds the correct slot when the correct label is available', () => {
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).not.toThrow();
        expect(mocks.C_OpenSession).toHaveBeenCalledWith(slot1, pkcs11js.CKF_SERIAL_SESSION);
    });

    it('defaults to a CKU_USER if none provided', () => {
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).not.toThrow();
        expect(mocks.C_Login).toHaveBeenCalledWith(mockSession, pkcs11js.CKU_USER, hsmOptions.pin);
    });

    it('uses usertype if provided', () => {
        const hsmOptionsWithUserType = {
            label: 'ForFabric',
            pin: '98765432',
            identifier: 'id',
            userType: 100,
        };

        expect(() => hsmSignerFactory.newSigner(hsmOptionsWithUserType)).not.toThrow();
        expect(mocks.C_Login).toHaveBeenCalledWith(mockSession, hsmOptionsWithUserType.userType, hsmOptions.pin);
    });

    it('throws if pkcs11 open session throws an error', () => {
        mocks.C_OpenSession.mockImplementation(() => {
            throw new Error('Some Error');
        });
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).toThrow('Some Error');
    });

    it('throws if pkcs11 login throws an error', () => {
        mocks.C_Login.mockImplementation(() => {
            throw new Error('Some Error');
        });
        mocks.C_GetSlotList.mockReturnValue([slot1, slot2]);
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).toThrow('Some Error');
        expect(mocks.C_CloseSession).toHaveBeenCalledWith(mockSession);
    });

    it('Ignores already logged in errors at login time', () => {
        const alreadyLoggedInError: Pkcs11Error = {
            code: pkcs11js.CKR_USER_ALREADY_LOGGED_IN,
            message: 'CKR_USER_ALREADY_LOGGED_IN',
            nativeStack: '[Native]',
            method: 'C_Login',
            name: 'error',
        };
        mocks.C_Login.mockImplementation(() => {
            throw alreadyLoggedInError;
        });
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).not.toThrow();
        expect(mocks.C_CloseSession).not.toHaveBeenCalled();
    });

    it('throws and calls find final if it cannot find the HSM object', () => {
        mocks.C_FindObjects.mockReturnValue([]);
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).toThrow('Unable to find object in HSM with ID id');
        expect(mocks.C_FindObjectsFinal).toHaveBeenCalled();
        expect(mocks.C_CloseSession).toHaveBeenCalledWith(mockSession);
    });

    it('finds the HSM object if it exists', () => {
        const signer = hsmSignerFactory.newSigner(hsmOptions);
        expect(signer).toBeDefined();

        const expectedTemplate = [
            { type: pkcs11js.CKA_ID, value: hsmOptions.identifier },
            { type: pkcs11js.CKA_CLASS, value: pkcs11js.CKO_PRIVATE_KEY },
            { type: pkcs11js.CKA_KEY_TYPE, value: pkcs11js.CKK_EC },
        ];

        expect(mocks.C_FindObjectsInit).toHaveBeenCalledWith(mockSession, expect.arrayContaining(expectedTemplate));
        expect(mocks.C_FindObjects).toHaveBeenCalledWith(mockSession, 1);
        expect(mocks.C_FindObjects).toHaveBeenCalledWith(mockSession, 1);
    });

    it('signs using the HSM', async () => {
        const message = Buffer.from('A quick brown fox jumps over the lazy dog');
        const digest = createHash('sha256').update(message).digest();

        const { signer } = hsmSignerFactory.newSigner(hsmOptions);
        const signature = await signer(digest);

        const valid = p256.verify(signature, digest, publicKey);
        expect(valid).toBe(true);

        expect(mocks.C_SignInit).toHaveBeenCalledWith(
            mockSession,
            { mechanism: pkcs11js.CKM_ECDSA },
            mockPrivateKeyHandle,
        );
        expect(mocks.C_SignAsync).toHaveBeenCalledWith(mockSession, digest, expect.anything());
    });

    it('can be closed', () => {
        const { close } = hsmSignerFactory.newSigner(hsmOptions);
        expect(() => {
            close();
        }).not.toThrow();
    });
});
