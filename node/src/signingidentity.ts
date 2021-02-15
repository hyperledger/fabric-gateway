/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Hash } from './hash/hash';
import { sha256 } from './hash/hashes';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';
import { msp } from './protos/protos';

const undefinedSigner: Signer = () => {
    throw new Error('No signing implementation');
}

export interface SigningIdentityOptions {
    identity: Identity;
    signer?: Signer;
    hash?: Hash;
}

export class SigningIdentity {
    readonly #identity: Identity;
    readonly #creator: Uint8Array;
    readonly #hash: Hash;
    readonly #sign: Signer;

    constructor(options: SigningIdentityOptions) {
        this.#identity = {
            mspId: options.identity.mspId,
            credentials: Uint8Array.from(options.identity.credentials)
        };

        this.#creator = msp.SerializedIdentity.encode({
            mspid: options.identity.mspId,
            id_bytes: options.identity.credentials
        }).finish();

        this.#hash = options.hash || sha256;
        this.#sign = options.signer || undefinedSigner;
    }

    getIdentity(): Identity {
        return {
            mspId: this.#identity.mspId,
            credentials: Uint8Array.from(this.#identity.credentials)
        };
    }

    getCreator(): Uint8Array {
        return Uint8Array.from(this.#creator);
    }

    hash(message: Uint8Array): Uint8Array {
        return this.#hash(message);
    }

    async sign(digest: Uint8Array): Promise<Uint8Array> {
        return this.#sign(digest);
    }
}
