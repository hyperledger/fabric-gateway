/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { TextDecoder, inspect } from 'util';

export function bytesAsString(bytes?: Uint8Array): string {
    return new TextDecoder().decode(bytes);
}

export function toString(thing: unknown): string {
    return inspect(thing);
}

export function toError(err: unknown): Error {
    if (err instanceof Error) {
        return err;
    }
    return new Error(toString(err));
}

export function assertDefined<T>(value: T | undefined, property: string): T {
    if (value === undefined) {
        throw new Error(`Bad step sequence: ${property} not defined`);
    }
    return value;
}
