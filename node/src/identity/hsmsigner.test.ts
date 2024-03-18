/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { p256 } from '@noble/curves/p256';
import { createHash } from 'node:crypto';
import { Mechanism, Pkcs11Error, SessionInfo, SlotInfo, Template, TokenInfo } from 'pkcs11js';
import { HSMSignerOptions } from './hsmsigner';
import { newHSMSignerFactory } from './signers';

const CKO_PRIVATE_KEY = 179;
const CKA_ID = 54;
const CKA_CLASS = 67;
const CKA_KEY_TYPE = 6;
const CKK_EC = 87;
const CKM_ECDSA = 532;
const CKF_SERIAL_SESSION = 24;
const CKU_USER = 72;
const CKR_USER_ALREADY_LOGGED_IN = 256;

const hsmOptions: HSMSignerOptions = {
    label: 'ForFabric',
    pin: '98765432',
    identifier: 'id',
};

// eslint-disable @typescript-eslint/no-unused-vars

const pkcs11Stub = {
    load: (): void => {
        return;
    },
    C_Initialize: (): void => {
        return;
    },
    C_GetInfo: (): string => 'Info',
    C_GetSlotList: (): Buffer[] => [],
    C_GetTokenInfo: (slot: Buffer): TokenInfo | null => null,
    C_GetSlotInfo: (slot: Buffer): SlotInfo | string => slot.toString(),
    C_GetMechanismList: (slot: Buffer): string[] => ['ECDSA'],
    C_OpenSession: (): void => {
        return;
    },
    C_GetSessionInfo: (): SessionInfo | void => {
        return;
    },
    C_Login: (): void => {
        return;
    },
    C_Logout: (session: Buffer): void => {
        return;
    },
    C_CloseSession: (): void => {
        return;
    },
    C_Finalize: (): void => {
        return;
    },
    C_FindObjectsInit: (session: Buffer, template: Template): void => {
        return;
    },
    C_FindObjects: (session: Buffer, limit: number): Buffer[] => {
        return [];
    },
    C_FindObjectsFinal: (session: Buffer): void => {
        return;
    },
    C_SignInit: (session: Buffer, mechanism: Mechanism, key: Buffer): void => {
        return;
    },
    C_Sign: (session: Buffer, digest: Buffer, store: Buffer): Buffer => {
        return digest;
    },
};

const resetPkcs11Stub: () => void = () => {
    pkcs11Stub.load = (): void => {
        return;
    };
    pkcs11Stub.C_Initialize = (): void => {
        return;
    };
    pkcs11Stub.C_GetInfo = (): string => 'Info';
    pkcs11Stub.C_GetSlotList = (): Buffer[] => [];
    pkcs11Stub.C_GetTokenInfo = (slot: Buffer): TokenInfo | null => null;
    pkcs11Stub.C_GetSlotInfo = (slot: Buffer): SlotInfo | string => slot.toString();
    pkcs11Stub.C_GetMechanismList = (slot: Buffer): string[] => ['ECDSA'];
    pkcs11Stub.C_OpenSession = (): void => {
        return;
    };
    pkcs11Stub.C_GetSessionInfo = (): void => {
        return;
    };
    pkcs11Stub.C_Login = (): void => {
        return;
    };
    pkcs11Stub.C_Logout = (session: Buffer): void => {
        return;
    };
    pkcs11Stub.C_CloseSession = (): void => {
        return;
    };
    pkcs11Stub.C_Finalize = (): void => {
        return;
    };
    pkcs11Stub.C_FindObjectsInit = (session: Buffer, template: Template): void => {
        return;
    };
    pkcs11Stub.C_FindObjects = (session: Buffer, limit: number): Buffer[] => {
        return [];
    };
    pkcs11Stub.C_FindObjectsFinal = (session: Buffer): void => {
        return;
    };
    pkcs11Stub.C_SignInit = (session: Buffer, mechanism: Mechanism, key: Buffer): void => {
        return;
    };
    pkcs11Stub.C_Sign = (session: Buffer, digest: Buffer, store: Buffer): Buffer => {
        return Buffer.from(digest);
    };
};

// eslint-enable @typescript-eslint/no-unused-vars

jest.mock('pkcs11js', () => {
    class PKCS11 {
        constructor() {
            return pkcs11Stub;
        }
    }

    // These are defined with random meaningless but unique values which have to be replicated because of jest
    const CKO_PRIVATE_KEY = 179;
    const CKA_ID = 54;
    const CKA_CLASS = 67;
    const CKA_KEY_TYPE = 6;
    const CKK_EC = 87;
    const CKM_ECDSA = 532;
    const CKF_SERIAL_SESSION = 24;
    const CKU_USER = 72;
    const CKR_USER_ALREADY_LOGGED_IN = 256;

    const exports = {
        PKCS11,
        CKO_PRIVATE_KEY,
        CKA_ID,
        CKA_CLASS,
        CKA_KEY_TYPE,
        CKK_EC,
        CKM_ECDSA,
        CKF_SERIAL_SESSION,
        CKU_USER,
        CKR_USER_ALREADY_LOGGED_IN,
    };
    return exports;
});

describe('when creating or disposing of an HSM Signer Factory', () => {
    beforeEach(() => {
        resetPkcs11Stub();
    });

    it('throws if library option is not valid', () => {
        pkcs11Stub.C_Initialize = () => {
            throw new Error('Some Error');
        };
        expect(() => newHSMSignerFactory('somelibrary')).toThrow('Some Error');

        expect(() => newHSMSignerFactory('')).toThrow('library must be provided');
    });

    it('can be disposed', () => {
        const hsmSignerFactory = newHSMSignerFactory('somelibrary');
        expect(() => hsmSignerFactory.dispose()).not.toThrow();
    });
});

describe('When using an HSM Signer', () => {
    const slot1 = Buffer.from('1234');
    const slot2 = Buffer.from('5678');
    const mockTokenInfo = (slot: Buffer): TokenInfo => {
        if (slot === slot1) {
            return { label: 'ForFabric' } as TokenInfo;
        }
        return { label: 'someLabel' } as TokenInfo;
    };

    const mockSession = Buffer.from('mockSession');
    const mockPrivateKeyHandle = Buffer.from('someobject');

    const hsmSignerFactory = newHSMSignerFactory('somelibrary');

    const privateKey = p256.utils.randomPrivateKey();
    const publicKey = p256.getPublicKey(privateKey);

    beforeEach(() => {
        resetPkcs11Stub();
        pkcs11Stub.C_GetTokenInfo = mockTokenInfo;
        pkcs11Stub.C_GetSlotList = () => [slot1, slot2];
        pkcs11Stub.C_OpenSession = () => {
            return mockSession;
        };
        pkcs11Stub.C_FindObjectsInit = jest.fn();
        pkcs11Stub.C_FindObjectsFinal = jest.fn();
        pkcs11Stub.C_FindObjects = jest.fn(() => {
            return [mockPrivateKeyHandle];
        });
        pkcs11Stub.C_SignInit = jest.fn();
        pkcs11Stub.C_Sign = jest.fn((session, digest, buffer) => {
            const signature = p256.sign(digest, privateKey).toCompactRawBytes();
            signature.forEach((b, i) => buffer.writeUInt8(b, i));
            // Return buffer of exactly signature length regardless of supplied buffer size
            return buffer.subarray(0, signature.length);
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
        pkcs11Stub.C_GetSlotList = () => [];
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
        pkcs11Stub.C_OpenSession = jest.fn();
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).not.toThrow();
        expect(pkcs11Stub.C_OpenSession).toHaveBeenCalledWith(slot1, CKF_SERIAL_SESSION);
    });

    it('defaults to a CKU_USER if none provided', () => {
        pkcs11Stub.C_Login = jest.fn();
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).not.toThrow();
        expect(pkcs11Stub.C_Login).toHaveBeenCalledWith(mockSession, CKU_USER, hsmOptions.pin);
    });

    it('uses usertype if provided', () => {
        const hsmOptionsWithUserType: HSMSignerOptions = {
            label: 'ForFabric',
            pin: '98765432',
            identifier: 'id',
            userType: 100,
        };

        pkcs11Stub.C_Login = jest.fn();
        expect(() => hsmSignerFactory.newSigner(hsmOptionsWithUserType)).not.toThrow();
        expect(pkcs11Stub.C_Login).toHaveBeenCalledWith(mockSession, hsmOptionsWithUserType.userType, hsmOptions.pin);
    });

    it('throws if pkcs11 open session throws an error', () => {
        pkcs11Stub.C_OpenSession = () => {
            throw new Error('Some Error');
        };
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).toThrow('Some Error');
    });

    it('throws if pkcs11 login throws an error', () => {
        pkcs11Stub.C_Login = () => {
            throw new Error('Some Error');
        };
        pkcs11Stub.C_CloseSession = jest.fn();
        pkcs11Stub.C_GetSlotList = () => [slot1, slot2];
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).toThrow('Some Error');
        expect(pkcs11Stub.C_CloseSession).toHaveBeenCalledWith(mockSession);
    });

    it('Ignores already logged in errors at login time', () => {
        pkcs11Stub.C_CloseSession = jest.fn();
        const alreadyLoggedInError: Pkcs11Error = {
            code: CKR_USER_ALREADY_LOGGED_IN,
            message: 'CKR_USER_ALREADY_LOGGED_IN',
            nativeStack: '[Native]',
            method: 'C_Login',
            name: 'error',
        };
        pkcs11Stub.C_Login = () => {
            throw alreadyLoggedInError;
        };
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).not.toThrow();
        expect(pkcs11Stub.C_CloseSession).not.toHaveBeenCalled();
    });

    it('throws and calls find final if it cannot find the HSM object', () => {
        pkcs11Stub.C_CloseSession = jest.fn();
        pkcs11Stub.C_FindObjects = jest.fn(() => {
            return [];
        });
        expect(() => hsmSignerFactory.newSigner(hsmOptions)).toThrow('Unable to find object in HSM with ID id');
        expect(pkcs11Stub.C_FindObjectsFinal).toHaveBeenCalled();
        expect(pkcs11Stub.C_CloseSession).toHaveBeenCalledWith(mockSession);
    });

    it('finds the HSM object if it exists', () => {
        const signer = hsmSignerFactory.newSigner(hsmOptions);
        expect(signer).toBeDefined();

        const expectedTemplate = [
            { type: CKA_ID, value: hsmOptions.identifier },
            { type: CKA_CLASS, value: CKO_PRIVATE_KEY },
            { type: CKA_KEY_TYPE, value: CKK_EC },
        ];

        expect(pkcs11Stub.C_FindObjectsInit).toHaveBeenCalledWith(
            mockSession,
            expect.arrayContaining(expectedTemplate),
        );
        expect(pkcs11Stub.C_FindObjects).toHaveBeenCalledWith(mockSession, 1);
        expect(pkcs11Stub.C_FindObjects).toHaveBeenCalledWith(mockSession, 1);
    });

    it('signs using the HSM', async () => {
        const message = Buffer.from('A quick brown fox jumps over the lazy dog');
        const digest = createHash('sha256').update(message).digest();

        const { signer } = hsmSignerFactory.newSigner(hsmOptions);
        const signature = await signer(digest);

        const valid = p256.verify(signature, digest, publicKey);
        expect(valid).toBe(true);

        expect(pkcs11Stub.C_SignInit).toHaveBeenCalledWith(mockSession, { mechanism: CKM_ECDSA }, mockPrivateKeyHandle);
        expect(pkcs11Stub.C_Sign).toHaveBeenCalledWith(mockSession, digest, expect.anything());
    });

    it('can be closed', () => {
        const { close } = hsmSignerFactory.newSigner(hsmOptions);
        expect(() => close()).not.toThrow();
    });
});
