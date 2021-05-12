/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as crypto from 'crypto';
import { Hash } from './hash';

/**
 * SHA256 hash the supplied message bytes to create a digest for signing.
 */
export const sha256: Hash = (message) => {
    return crypto.createHash('sha256').update(message).digest();
}
