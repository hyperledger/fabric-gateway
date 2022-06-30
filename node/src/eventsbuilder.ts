/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { orderer } from '@hyperledger/fabric-protos';
import { Checkpoint } from './checkpointer';

/**
 * Options used when requesting events.
 *
 * If both a start block and checkpoint are specified, and the checkpoint has a valid position set, the checkpoint
 * position is used and the specified start block is ignored. If the checkpoint is unset then the start block is used.
 *
 * If no start position is specified, eventing begins from the next committed block.
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

    getStartPosition(): orderer.SeekPosition {
        const result = new orderer.SeekPosition();
        const startBlock = this.#options.checkpoint?.getBlockNumber() ?? this.#options.startBlock;

        if (startBlock != undefined) {
            const specified = new orderer.SeekSpecified();

            specified.setNumber(Number(startBlock));
            result.setSpecified(specified);

            return result;
        }

        result.setNextCommit(new orderer.SeekNextCommit());
        return result;
    }
}
