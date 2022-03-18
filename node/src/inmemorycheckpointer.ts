/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Checkpointer } from './checkpointer';
import { ChaincodeEvent } from './chaincodeevent';

/**
 * In-memory checkpointer class used to persist checkpointer state in memory.
 */

export class InMemoryCheckPointer implements Checkpointer {

	#blockNumber?: bigint;
	#transactionID?: string;

	checkpointBlock(blockNumber: bigint): void {
		this.#blockNumber = blockNumber + BigInt(1);
		this.#transactionID = undefined;
	}

	checkpointTransaction(blockNumber: bigint, transactionId: string): void {
		this.#blockNumber = blockNumber;
		this.#transactionID = transactionId;
	}

	checkpointChaincodeEvent(event: ChaincodeEvent): void {
		this.checkpointTransaction(event.blockNumber,event.transactionId);
	}

	getBlockNumber(): bigint | undefined {
		return this.#blockNumber;
	}

	getTransactionId(): string | undefined {
		return this.#transactionID;
	}
}
