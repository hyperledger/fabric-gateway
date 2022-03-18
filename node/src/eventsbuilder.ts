/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Checkpointer } from '.';
import { SeekNextCommit, SeekPosition, SeekSpecified } from './protos/orderer/ab_pb';

/**
 * Options used when requesting events.
 */
export interface EventsOptions {
    /**
     * Block number at which to start reading events.
     */
    startBlock?: bigint;

    checkpointer?: Checkpointer;
}

export class EventsBuilder {
    readonly #options: Readonly<EventsOptions>;

    constructor(options: Readonly<EventsOptions>) {
        this.#options = options;
    }

    getStartPosition(): SeekPosition {
        const result = new SeekPosition();
        const specified = new SeekSpecified();
        const startBlock = this.#options.startBlock;
        const checkPointer = this.#options.checkpointer;

        if (checkPointer) {
            let blockNumber = checkPointer.getBlockNumber();
            if (blockNumber != undefined) {
                specified.setNumber(Number(blockNumber));
                result.setSpecified(specified);

                return result;
            }
        }
        if (startBlock != undefined) {
            specified.setNumber(Number(startBlock));
            result.setSpecified(specified);

            return result;
        }

        result.setNextCommit(new SeekNextCommit());
        return result;
    }
}
