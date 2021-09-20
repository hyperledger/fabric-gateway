/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

export interface ErrorFirstCallback<T> {
    (err: undefined, event: T): Promise<void>;
    (err: unknown, event: undefined): Promise<void>;
}
