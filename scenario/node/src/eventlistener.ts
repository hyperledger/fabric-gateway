/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

export interface EventListener<T> {
    next(): Promise<T>;
    close(): void;
}
