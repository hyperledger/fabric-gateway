/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Checkpointer } from './checkpointer';
import { FileCheckPointer } from './filecheckpointer';
import { InMemoryCheckPointer } from './inmemorycheckpointer';

/**
 * Create a checkpointer that uses the specified file to store persistent state.
 * @param path - Path to a file holding persistent checkpoint state.
 */
export async function file(path: string): Promise<Checkpointer> {
    const filecheckpointer = new FileCheckPointer(path);
    await filecheckpointer.init();
    return filecheckpointer;
}

/**
 * Create a checkpointer that stores its state in memory only.
 */
export function inMemory(): Checkpointer {
    return new InMemoryCheckPointer();
}
