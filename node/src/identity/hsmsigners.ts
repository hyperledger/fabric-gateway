/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as pkcs11js from 'pkcs11js';
import * as elliptic from 'elliptic';
import BN = require("bn.js");
import ecSignature = require('elliptic/lib/elliptic/ec/signature');
import { Signer } from './signer';

type DER = string | number[];

export interface HSMSignerOptions {
    /**
     * The label associated with the token for the slot
     */
    label: string;

    /**
     * The pin for the slot identified by the label
     */
    pin: string;

    /**
     * Identifier. The CKA_ID assigned to the HSM object
     */
    identifier: string | Buffer;

    /**
     * Optional user type for the HSM. If not specified it defaults to CKU_USER
     */
    userType?: number;
}

export interface HSMSigner {
    signer: Signer;
    close: () => void;
}
/**
 * Create an HSM Signer factory. You should only ever call this once within your application
 */
export function newHSMSignerFactory(library: string): HSMSignerFactory {
    if (!library || library.trim() === '') {
        throw new Error('library must be provided');
    }

    return new HSMSignerFactory(library);
}

/**
 * Factory to be able to create HSM Signers.
 */
export class HSMSignerFactory {

    #pkcs11: pkcs11js.PKCS11;

    constructor(library: string) {
        this.#pkcs11 = new pkcs11js.PKCS11();
        this.#pkcs11.load(library);
        this.#pkcs11.C_Initialize();
    }

    /**
     * disposes of the factory when it and HSM signers are not required anymore
     */
    public dispose(): void {
        this.#pkcs11.C_Finalize();
    }

    /**
     * Create a new HSM signing implementation based on provided HSM options.
     *
     * This returns an object with 2 properties
     * - signer which is the signer function
     * - close which is a close function to close the signer when it's not required anymore
     *
     * @param hsmSignerOptions - The HSM signer options
     * @returns an HSM Signer implementation
     */
    public newSigner(hsmSignerOptions: HSMSignerOptions): HSMSigner {
        if (!hsmSignerOptions.label || hsmSignerOptions.label.trim() === '') {
            throw new Error('label property must be provided');
        }

        if (!hsmSignerOptions.pin || hsmSignerOptions.pin.trim() === '') {
            throw new Error('pin property must be provided');
        }

        if (!hsmSignerOptions.identifier || hsmSignerOptions.identifier.toString().trim() === '') {
            throw new Error('identifier property must be provided');
        }

        if (!hsmSignerOptions.userType) {
            hsmSignerOptions.userType = pkcs11js.CKU_USER;
        }

        const supportedKeySize = 256;
        const pkcs11 = this.#pkcs11;
        const slot = this.findSlotForLabel(hsmSignerOptions.label);
        const session = pkcs11.C_OpenSession(slot, pkcs11js.CKF_SERIAL_SESSION);
        pkcs11.C_Login(session, hsmSignerOptions.userType, hsmSignerOptions.pin);
        const privateKeyHandle = this.findObjectInHSM(session, pkcs11js.CKO_PRIVATE_KEY, hsmSignerOptions.identifier);

        const definedCurves = elliptic.curves as unknown as { [key: string]: elliptic.curves.PresetCurve };
        const ecdsaCurve = definedCurves[`p${supportedKeySize}`];

        // currently the only supported curve is p256 and it will always have an 'n' value
        const curveBigNum = ecdsaCurve.n!; // eslint-disable-line @typescript-eslint/no-non-null-assertion
        const halfOrder = curveBigNum.shrn(1);

        const close = ():void => {
            pkcs11.C_Logout(session);
            pkcs11.C_CloseSession(session);
        }

        const signer: Signer = async (digest: Uint8Array) => {
            pkcs11.C_SignInit(session, { mechanism: pkcs11js.CKM_ECDSA }, privateKeyHandle);
            const sig = pkcs11.C_Sign(session, Buffer.from(digest), Buffer.alloc(supportedKeySize));

            const r = new BN(sig.slice(0, sig.length / 2).toString('hex'), 16);
            let s = new BN(sig.slice(sig.length / 2).toString('hex'), 16);

            if (s.cmp(halfOrder) === 1) {
                s = curveBigNum.sub(s);
            }

            const signatureInput: elliptic.SignatureInput = {
                r,
                s
            }

            const der = new ecSignature(signatureInput).toDER() as DER; // eslint-disable-line
            return Promise.resolve(Buffer.from(der));
        }

        return {signer, close};
    }

    private findSlotForLabel(pkcs11Label: string): Buffer {
        const slots = this.#pkcs11.C_GetSlotList(true);

        if (!slots || slots.length === 0) {
            throw new Error('No pkcs11 slots can be found');
        }

        let slot: Buffer | undefined;
        let tokenInfo: pkcs11js.TokenInfo;

        for (const slotToCheck of slots) {
            tokenInfo = this.#pkcs11.C_GetTokenInfo(slotToCheck);
            if (tokenInfo && tokenInfo.label && tokenInfo.label.trim() === pkcs11Label) {
                slot = slotToCheck;
                break;
            }
        }

        if (!slot) {
            throw new Error(`label ${pkcs11Label} cannot be found in the pkcs11 slot list`);
        }

        return slot;
    }

    private findObjectInHSM(session: Buffer, keytype: number, identifier: string | Buffer): Buffer {
        const pkcs11Template: pkcs11js.Template = [
            { type: pkcs11js.CKA_ID, value: identifier },
            { type: pkcs11js.CKA_CLASS, value: keytype },
            { type: pkcs11js.CKA_KEY_TYPE, value: pkcs11js.CKK_EC }
        ]
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
