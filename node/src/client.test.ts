/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable jest/no-export */

import { GatewayClient } from './client';
import { gateway } from './protos/protos';

export interface MockGatewayClient extends GatewayClient {
    endorse: jest.Mock<Promise<gateway.IEndorseResponse>, gateway.IEndorseRequest[]>,
    evaluate: jest.Mock<Promise<gateway.IEvaluateResponse>, gateway.IEvaluateRequest[]>,
    submit: jest.Mock<Promise<gateway.ISubmitResponse>, gateway.ISubmitRequest[]>,
    commitStatus: jest.Mock<Promise<gateway.ICommitStatusResponse>, gateway.ISignedCommitStatusRequest[]>,
    chaincodeEvents: jest.Mock<AsyncIterable<gateway.IChaincodeEventsResponse>, gateway.ISignedChaincodeEventsRequest[]>,
}

export function newMockGatewayClient(): MockGatewayClient {
    return {
        endorse: jest.fn(undefined),
        evaluate: jest.fn(undefined),
        submit: jest.fn(undefined),
        commitStatus: jest.fn(undefined),
        chaincodeEvents: jest.fn(() => {
            return {
                async* [Symbol.asyncIterator]() {
                    // Nothing
                }
            };
        }),
    };
}

describe('client', () => {
    it('dummy test to avoid missing test error', () => {
        expect(newMockGatewayClient()).toBeDefined();
    });
});
