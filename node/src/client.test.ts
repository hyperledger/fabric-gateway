/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable jest/no-export */

import * as grpc from '@grpc/grpc-js';
import { UnaryCallback } from '@grpc/grpc-js/build/src/client';
import { GatewayError } from './gatewayerror';
import { CloseableAsyncIterable, GatewayClient, newGatewayClient } from './client';
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
    describe('throws GatewayError on gRPC error', () => {
        let grpcError: grpc.ServiceError;
        let grpcClient: grpc.Client;
        let gatewayClient: GatewayClient;
    
        beforeEach(() => {
            grpcError = Object.assign(new Error('GRPC_STATUS_ERROR'), {
                code: grpc.status.UNAVAILABLE,
                details: 'GRPC_DETAILS',
                metadata: new grpc.Metadata(),
            });

            grpcClient = {
                makeUnaryRequest: (method: unknown, serialize: unknown, deserialize: unknown, argument: unknown, callback: UnaryCallback<unknown>) => {
                    setImmediate(() => callback(grpcError, undefined));
                },
            } as grpc.Client;

            gatewayClient = newGatewayClient(grpcClient);
        });

        const tests: {
            name: string,
            run: () => Promise<unknown>,
        }[] = [
            { name: 'endorse', run: () => gatewayClient.endorse({} as EndorseRequest) },
            { name: 'evaluate', run: () => gatewayClient.evaluate({} as EvaluateRequest) },
            { name: 'submit', run: () => gatewayClient.submit({} as SubmitRequest) },
            { name: 'commitStatus', run: () => gatewayClient.commitStatus({} as SignedCommitStatusRequest) },
        ];

        tests.forEach(test => {
            it(`${test.name}`, async () => {
                const t = test.run();
                await expect(t).rejects.toThrow(grpcError.message);
                await expect(t).rejects.toThrow(GatewayError);
            });
        });
    });
});
