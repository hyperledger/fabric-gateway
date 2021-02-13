/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as crypto from 'crypto';

export function sha256(message: Uint8Array): Uint8Array {
    return crypto.createHash('sha256').update(message).digest();
}
