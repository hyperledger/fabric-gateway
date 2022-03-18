/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Checkpoint } from './checkpointer';
import { SeekNextCommit, SeekPosition, SeekSpecified } from './protos/orderer/ab_pb';

/**
 * Options used when requesting events.
 */
export interface EventsOptions {
    /**
     * Block number at which to start reading events.
     */
    startBlock?: bigint;

    /**
     * Used to get checkpointed state.
     */
    checkpoint?: Checkpoint;
}

export class EventsBuilder {
    readonly #options: Readonly<EventsOptions>;

    constructor(options: Readonly<EventsOptions>) {
        this.#options = options;
    }

    getStartPosition(): SeekPosition {
        const result = new SeekPosition();
        const startBlock = this.#options.checkpoint?.getBlockNumber() ?? this.#options.startBlock;

        if (startBlock != undefined) {
            const specified = new SeekSpecified();

            specified.setNumber(Number(startBlock));
            result.setSpecified(specified);

            return result;
        }

        result.setNextCommit(new SeekNextCommit());
        return result;
    }
}
