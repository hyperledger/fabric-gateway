/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

export interface CheckPointer {
  checkpoint(blockNumber: bigint, transactionId?: string): Promise<void>;
  getBlockNumber(): Promise<bigint | undefined>;
  getTransactionIds(): Promise<string[]>;
}

export interface CheckPointerState {
  blockNumber?: bigint;
  transactionIDs: string[];
}
