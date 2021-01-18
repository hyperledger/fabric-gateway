/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as crypto from 'crypto';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';
import { msp } from './protos/protos';

const undefinedSigner: Signer = () => {
    throw new Error('No signing implementation');
}

export class SigningIdentity {
    readonly #identity: Identity;
    readonly #creator: Uint8Array;
    readonly #sign: Signer;

    constructor(identity: Identity, signer?: Signer) {
        this.#identity = {
            mspId: identity.mspId,
            credentials: Uint8Array.from(identity.credentials)
        };

        this.#creator = msp.SerializedIdentity.encode({
            mspid: identity.mspId,
            id_bytes: identity.credentials
        }).finish();

        this.#sign = signer || undefinedSigner;
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
        return crypto.createHash('sha256').update(message).digest();
    }

    async sign(digest: Uint8Array): Promise<Uint8Array> {
        return this.#sign(digest);
    }
}
