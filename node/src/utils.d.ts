/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

export interface Callback<T> {
    (err: unknown, event: undefined): Promise<void>;
    (err: undefined, event: T): Promise<void>;
}
