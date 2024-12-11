/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

declare const globalThis:
    | {
          crypto?: Partial<Crypto>;
      }
    | undefined;

const crypto: Partial<Crypto> = globalThis?.crypto ?? {};

const getRandomValues: <T extends ArrayBufferView>(array: T) => T = (
    crypto.getRandomValues ?? error('No globalThis.crypto.getRandomValues defined')
).bind(crypto);

export function randomBytes(size: number): Uint8Array {
    return getRandomValues(new Uint8Array(size));
}

const subtle: Partial<SubtleCrypto> = crypto.subtle ?? {};

const digest: (algorithm: AlgorithmIdentifier, data: BufferSource) => Promise<ArrayBuffer> = (
    subtle.digest ?? error('No globalThis.crypto.subtle.digest defined')
).bind(subtle);

export async function sha256(message: Uint8Array): Promise<Uint8Array> {
    const result = await digest('SHA-256', message);
    return new Uint8Array(result);
}

function error(message: string): () => never {
    return () => {
        throw new Error(message);
    };
}
