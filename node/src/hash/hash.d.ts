/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * A hashing implementation used to generate a digest from a supplied message.
 */
export type Hash = (message: Uint8Array) => Uint8Array;
