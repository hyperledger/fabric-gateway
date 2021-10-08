/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable jest/no-export */

import { CloseableAsyncIterable, GatewayClient } from './client';
import { ChaincodeEventsResponse, CommitStatusResponse, EndorseRequest, EndorseResponse, EvaluateRequest, EvaluateResponse, SignedChaincodeEventsRequest, SignedCommitStatusRequest, SubmitRequest, SubmitResponse } from './protos/gateway/gateway_pb';

export interface MockGatewayClient extends GatewayClient {
    endorse: jest.Mock<Promise<EndorseResponse>, EndorseRequest[]>,
    evaluate: jest.Mock<Promise<EvaluateResponse>, EvaluateRequest[]>,
    submit: jest.Mock<Promise<SubmitResponse>, SubmitRequest[]>,
    commitStatus: jest.Mock<Promise<CommitStatusResponse>, SignedCommitStatusRequest[]>,
    chaincodeEvents: jest.Mock<CloseableAsyncIterable<ChaincodeEventsResponse>, SignedChaincodeEventsRequest[]>,
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
                },
                close: jest.fn(),
            };
        }),
    };
}

describe('client', () => {
    it('dummy test to avoid missing test error', () => {
        expect(newMockGatewayClient()).toBeDefined();
    });
});
