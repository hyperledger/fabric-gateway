/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Mapping of a type with all properties required and non-null.
 */
export type Mandatory<T> = {
    [P in keyof T]-?: NonNullable<T[P]>
};

/**
 * Mapping of a type with named properties required and non-null.
 */
export type MandatoryProperties<T, K extends keyof T> = Omit<T, K> & Mandatory<Pick<T, K>>;
