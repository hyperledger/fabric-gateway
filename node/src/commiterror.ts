/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { peer } from '@hyperledger/fabric-protos';
import { Status, StatusNames } from './status';

/**
 * CommitError is thrown to indicate that a transaction committed with an unsuccessful status code.
 */
export class CommitError extends Error {
    /**
     * Transaction validation status code. The value corresponds to one of the values enumerated by {@link StatusCode}.
     */
    code: peer.TxValidationCodeMap[keyof peer.TxValidationCodeMap];

    /**
     * The ID of the transaction.
     */
    transactionId: string;

    constructor(properties: Readonly<Omit<CommitError, keyof Error> & Partial<Pick<Error, 'message'>>>) {
        super(properties.message);

        this.name = CommitError.name;
        this.code = properties.code;
        this.transactionId = properties.transactionId;
    }
}

export function newCommitError(status: Status): CommitError {
    return new CommitError({
        message: `Transaction ${status.transactionId} failed to commit with status code ${status.code} (${StatusNames[status.code]})`,
        code: status.code,
        transactionId: status.transactionId,
    });
}
