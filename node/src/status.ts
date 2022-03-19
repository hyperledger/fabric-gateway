/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { peer } from '@hyperledger/fabric-protos';

/**
 * Enumeration of transaction status codes.
 */
export const StatusCode = Object.freeze(peer.TxValidationCode) as { [P in keyof typeof peer.TxValidationCode]: typeof peer.TxValidationCode[P] };

export const StatusNames = Object.freeze(
    Object.fromEntries(
        Object.entries(StatusCode)
            .filter(([_, code]) => typeof code === 'number') // eslint-disable-line @typescript-eslint/no-unused-vars
            .map(([name, code]) => [code, name])
    ) as { [P in keyof typeof StatusCode as typeof StatusCode[P]]: P }
);

/**
 * Status of a transaction that is committed to the ledger.
 */
export interface Status {
    /**
     * Block number in which the transaction committed.
     */
    blockNumber: bigint;

    /**
     * Transaction validation status code. The value corresponds to one of the values enumerated by {@link StatusCode}.
     */
    code: peer.TxValidationCodeMap[keyof peer.TxValidationCodeMap];

    /**
     * `true` if the transaction committed successfully; otherwise `false`.
     */
    successful: boolean;

    /**
     * The ID of the transaction.
     */
    transactionId: string;
}
