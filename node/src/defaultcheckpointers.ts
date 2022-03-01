/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { FileCheckPointer } from "./filecheckpointer";
import { InmemoryCheckPointer } from "./inmemorycheckpointer";

export class DefaultCheckPointers {

  static async file(path: string):Promise<FileCheckPointer>{
    const filecheckpointer = new FileCheckPointer(path);
    await filecheckpointer.init();
    return filecheckpointer;
  }

  static inMemory():InmemoryCheckPointer{
    const inmemorycheckpointer = new InmemoryCheckPointer();
    return inmemorycheckpointer;
  }
}
