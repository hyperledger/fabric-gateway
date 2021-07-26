/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { TextDecoder } from 'util';

export function asString(bytes?: Uint8Array): string {
    return new TextDecoder().decode(bytes);
}

export function assertDefined<T>(value: T | undefined, property: string): T {
    if (value === undefined) {
        throw new Error(`Bad step sequence: ${property} not defined`);
    }
    return value;
}
