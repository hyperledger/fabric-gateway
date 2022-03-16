/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { SeekNextCommit, SeekPosition, SeekSpecified } from './protos/orderer/ab_pb';

/**
 * Options used when checkpointng events.
 */
export interface CheckpointerData {
    /**
     * Block number at which to start reading events.
     */
    startBlock?: bigint;
    /**
    * Tranasction ID of the last processed event .
    */
    afterTransactionID?: string;
}

/**
 * Options used when requesting events.
 */
export interface EventsOptions {
    /**
     * Block number at which to start reading events.
     */
    startBlock?: bigint;

    checkpointer?: CheckpointerData;
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
            if (checkPointer.startBlock != undefined) {
                specified.setNumber(Number(checkPointer.startBlock));
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
