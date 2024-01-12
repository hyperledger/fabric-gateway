/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ServiceError } from '@grpc/grpc-js';
import { ErrorDetail, GatewayError } from './gatewayerror';

/**
 * EndorseError is thrown when a failure occurs endorsing a transaction proposal.
 */
export class EndorseError extends GatewayError {
    /**
     * The ID of the transaction.
     */
    transactionId: string;

    constructor(
        properties: Readonly<{
            code: number;
            details: ErrorDetail[];
            cause: ServiceError;
            transactionId: string;
            message?: string;
        }>,
    ) {
        super(properties);

        this.name = EndorseError.name;
        this.transactionId = properties.transactionId;
    }
}
