/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { msp } from '@hyperledger/fabric-protos';
import { ConnectOptions } from './gateway';
import { Hash } from './hash/hash';
import { sha256 } from './hash/hashes';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';

export const undefinedSignerMessage = 'No signing implementation';

const undefinedSigner: Signer = () => {
    throw new Error(undefinedSignerMessage);
};

type SigningIdentityOptions = Pick<ConnectOptions, 'identity' | 'signer' | 'hash'>;

export class SigningIdentity {
    readonly #identity: Identity;
    readonly #creator: Uint8Array;
    readonly #hash: Hash;
    readonly #sign: Signer;

    constructor(options: Readonly<SigningIdentityOptions>) {
        this.#identity = {
            mspId: options.identity.mspId,
            credentials: Uint8Array.from(options.identity.credentials),
        };

        const serializedIdentity = new msp.SerializedIdentity();
        serializedIdentity.setMspid(options.identity.mspId);
        serializedIdentity.setIdBytes(options.identity.credentials);
        this.#creator = serializedIdentity.serializeBinary();

        this.#hash = options.hash ?? sha256;
        this.#sign = options.signer ?? undefinedSigner;
    }

    getIdentity(): Identity {
        return {
            mspId: this.#identity.mspId,
            credentials: Uint8Array.from(this.#identity.credentials),
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
