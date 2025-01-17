/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * A signing implementation used to generate digital signatures from a supplied message. Note that a complete message
 * is supplied, which might need to be hashed to create a digest before signing.
 * @param message - Complete message bytes.
 */
export type Signer = (message: Uint8Array) => Promise<Uint8Array>;
