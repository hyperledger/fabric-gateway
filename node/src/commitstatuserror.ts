/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayError } from './gatewayerror';

/**
 * CommitStatusError is thrown when a failure occurs obtaining the commit status of a transaction.
 */
export class CommitStatusError extends GatewayError {
    /**
     * The ID of the transaction.
     */
    transactionId: string;

    constructor(properties: Readonly<Omit<CommitStatusError, keyof Error> & Partial<Pick<Error, 'message'>>>) {
        super(properties);

        this.name = CommitStatusError.name;
        this.transactionId = properties.transactionId;
    }
}
