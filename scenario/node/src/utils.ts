/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { inspect, TextDecoder } from 'util';

const utf8Decoder = new TextDecoder();

export function bytesAsString(bytes?: Uint8Array): string {
    return utf8Decoder.decode(bytes);
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

export interface Constructor<T> {
    new (...args: never[]): T;
}

export function isInstanceOf<T>(o: unknown, type: Constructor<T>): o is T {
    return o instanceof type;
}
