/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { SerializedIdentity } from '@hyperledger/fabric-protos/lib/msp/identities_pb';
import { ConnectOptions } from './gateway';
import { Identity } from './identity';
import { Signer } from './signer';

type SigningIdentityOptions = Pick<ConnectOptions, 'identity' | 'signer'>;

export class SigningIdentity {
    readonly #identity: Identity;
    readonly #creator: Uint8Array;
    readonly #sign: Signer;

    constructor(options: Readonly<SigningIdentityOptions>) {
        this.#identity = {
            mspId: options.identity.mspId,
            credentials: Uint8Array.from(options.identity.credentials),
        };

        const serializedIdentity = new SerializedIdentity();
        serializedIdentity.setMspid(options.identity.mspId);
        serializedIdentity.setIdBytes(options.identity.credentials);
        this.#creator = serializedIdentity.serializeBinary();

        this.#sign = options.signer;
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

    sign(message: Uint8Array): Promise<Uint8Array> {
        return this.#sign(message);
    }
}
