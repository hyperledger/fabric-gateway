/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import BN from 'bn.js';
import * as elliptic from 'elliptic';
import * as pkcs11js from 'pkcs11js';
import { ecRawSignatureAsDer } from './asn1';
import { Signer } from './signer';

export interface HSMSignerOptions {
    /**
     * The label associated with the token for the slot.
     */
    label: string;

    /**
     * The pin for the slot identified by the label.
     */
    pin: string;

    /**
     * Identifier. The CKA_ID assigned to the HSM object.
     */
    identifier: string | Buffer;

    /**
     * Optional user type for the HSM. If not specified it defaults to CKU_USER.
     */
    userType?: number;
}

export interface HSMSigner {
    /**
     * HSM signer implementation.
     */
    signer: Signer;

    /**
     * Close the HSM session when the signer is no longer needed.
     */
    close: () => void;
}

/**
 * Factory to create HSM Signers.
 */
export interface HSMSignerFactory {
    /**
     * Create a new HSM signing implementation based on provided HSM options.
     *
     * This returns an object with two properties:
     * 1. the signer function.
     * 2. a close function to be called when the signer is no longer required.
     *
     * @param hsmSignerOptions - The HSM signer options.
     * @returns an HSM Signer implementation.
     */
    newSigner(hsmSignerOptions: HSMSignerOptions): HSMSigner;

    /**
     * Dispose of the factory when it, and any HSM signers created by it, are no longer required.
     */
    dispose(): void;
}

export class HSMSignerFactoryImpl implements HSMSignerFactory {
    readonly #pkcs11: pkcs11js.PKCS11;

    constructor(library: string) {
        this.#pkcs11 = new pkcs11js.PKCS11();
        this.#pkcs11.load(library);
        this.#pkcs11.C_Initialize();
    }

    dispose(): void {
        this.#pkcs11.C_Finalize();
    }

    newSigner(hsmSignerOptions: Readonly<HSMSignerOptions>): HSMSigner {
        const options = sanitizeOptions(hsmSignerOptions);

        const supportedKeySize = 256;
        const pkcs11 = this.#pkcs11;
        const slot = this.#findSlotForLabel(options.label);
        const session = pkcs11.C_OpenSession(slot, pkcs11js.CKF_SERIAL_SESSION);


        let privateKeyHandle: Buffer;
        try {
            this.#login(session, options.userType, options.pin);
            privateKeyHandle = this.#findObjectInHSM(session, pkcs11js.CKO_PRIVATE_KEY, options.identifier);
        } catch (err) {
            pkcs11.C_CloseSession(session);
            throw err;
        }

        const definedCurves = elliptic.curves as unknown as Record<string, elliptic.curves.PresetCurve>;
        const ecdsaCurve = definedCurves[`p${supportedKeySize}`];

        // currently the only supported curve is p256 and it will always have an 'n' value
        const curveBigNum = ecdsaCurve.n!; // eslint-disable-line @typescript-eslint/no-non-null-assertion
        const halfOrder = curveBigNum.shrn(1);

        return {
            signer: (digest) => {
                pkcs11.C_SignInit(session, { mechanism: pkcs11js.CKM_ECDSA }, privateKeyHandle);
                const sig = pkcs11.C_Sign(session, Buffer.from(digest), Buffer.alloc(supportedKeySize));

                const r = new BN(sig.slice(0, sig.length / 2).toString('hex'), 16);
                let s = new BN(sig.slice(sig.length / 2).toString('hex'), 16);

                if (s.cmp(halfOrder) === 1) {
                    s = curveBigNum.sub(s);
                }

                const signature = new Uint8Array(ecRawSignatureAsDer(r, s));
                return Promise.resolve(signature);
            },
            close: () => pkcs11.C_CloseSession(session),
        };
    }

    #findSlotForLabel(pkcs11Label: string): Buffer {
        const slots = this.#pkcs11.C_GetSlotList(true);

        if (!slots || slots.length === 0) {
            throw new Error('No pkcs11 slots can be found');
        }

        const slot = slots.find(slotToCheck => {
            const tokenInfo = this.#pkcs11.C_GetTokenInfo(slotToCheck);
            return tokenInfo?.label?.trim() === pkcs11Label;
        });

        if (!slot) {
            throw new Error(`label ${pkcs11Label} cannot be found in the pkcs11 slot list`);
        }

        return slot;
    }


    #login(session: Buffer, userType: number, pin: string): void {
        try {
            this.#pkcs11.C_Login(session, userType, pin);
        } catch (err) {
            const pkcs11err = err as pkcs11js.Pkcs11Error;
            if (pkcs11err.code !== pkcs11js.CKR_USER_ALREADY_LOGGED_IN) {
                throw err;
            }
        }
    }

    #findObjectInHSM(session: Buffer, keytype: number, identifier: string | Buffer): Buffer {
        const pkcs11Template: pkcs11js.Template = [
            { type: pkcs11js.CKA_ID, value: identifier },
            { type: pkcs11js.CKA_CLASS, value: keytype },
            { type: pkcs11js.CKA_KEY_TYPE, value: pkcs11js.CKK_EC }
        ];
        this.#pkcs11.C_FindObjectsInit(session, pkcs11Template);

        const hsmObjects = this.#pkcs11.C_FindObjects(session, 1);

        if (!hsmObjects || hsmObjects.length === 0) {
            this.#pkcs11.C_FindObjectsFinal(session);
            throw new Error(`Unable to find object in HSM with ID ${identifier.toString()}`);
        }

        this.#pkcs11.C_FindObjectsFinal(session);

        return hsmObjects[0];
    }
}

function sanitizeOptions(hsmSignerOptions: HSMSignerOptions): Required<HSMSignerOptions> {
    const options = Object.assign({
        userType: pkcs11js.CKU_USER,
    }, hsmSignerOptions);

    assertNotEmpty(options.label, 'label');
    assertNotEmpty(options.pin, 'pin');
    assertNotEmpty(options.identifier, 'identifier');

    return options;
}

function assertNotEmpty(property: string | Buffer | undefined, name: string): void {
    if (!property || property.toString().trim().length === 0) {
        throw new Error(`${name} property must be provided`);
    }
}
