/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable jest/no-export */

import * as grpc from '@grpc/grpc-js';
import { CloseableAsyncIterable, GatewayClient, newGatewayClient } from './client';
import { GatewayError } from './gatewayerror';
import { Envelope } from './protos/common/common_pb';
import { ChaincodeEventsResponse, CommitStatusResponse, EndorseRequest, EndorseResponse, EvaluateRequest, EvaluateResponse, SignedChaincodeEventsRequest, SignedCommitStatusRequest, SubmitRequest, SubmitResponse } from './protos/gateway/gateway_pb';
import { DeliverResponse } from './protos/peer/events_pb';
import { MockGatewayGrpcClient } from './testutils.test';

describe('client', () => {
    describe('throws GatewayError on gRPC error', () => {
        const grpcError: grpc.ServiceError = Object.assign(new Error('GRPC_STATUS_ERROR'), {
            code: grpc.status.UNAVAILABLE,
            details: 'GRPC_DETAILS',
            metadata: new grpc.Metadata(),
        });

        let grpcClient: MockGatewayGrpcClient;
        let gatewayClient: GatewayClient;

        beforeEach(() => {
            grpcClient = new MockGatewayGrpcClient();
            gatewayClient = newGatewayClient(grpcClient, {});
        });

        it('evaluate', async () => {
            grpcClient.mockEvaluateError(grpcError);

            const t: () => Promise<EvaluateResponse> = () => gatewayClient.evaluate(new EvaluateRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('endorse', async () => {
            grpcClient.mockEndorseError(grpcError);

            const t: () => Promise<EndorseResponse> = () => gatewayClient.endorse(new EndorseRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('submit', async () => {
            grpcClient.mockSubmitError(grpcError);

            const t: () => Promise<SubmitResponse> = () => gatewayClient.submit(new SubmitRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('commit status', async () => {
            grpcClient.mockCommitStatusError(grpcError);

            const t: () => Promise<CommitStatusResponse> = () => gatewayClient.commitStatus(new SignedCommitStatusRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('chaincode events', () => {
            grpcClient.mockChaincodeEventsError(grpcError);

            const t: () => CloseableAsyncIterable<ChaincodeEventsResponse> = () => gatewayClient.chaincodeEvents(new SignedChaincodeEventsRequest());

            expect(t).toThrow(grpcError.message);
            expect(t).toThrow(GatewayError);
        });

        it('block events', () => {
            grpcClient.mockBlockEventsError(grpcError);

            const t: () => CloseableAsyncIterable<DeliverResponse> = () => gatewayClient.blockEvents(new Envelope());

            expect(t).toThrow(grpcError.message);
            expect(t).toThrow(GatewayError);
        });

        it('filtered block events', () => {
            grpcClient.mockFilteredBlockEventsError(grpcError);

            const t: () => CloseableAsyncIterable<DeliverResponse> = () => gatewayClient.filteredBlockEvents(new Envelope());

            expect(t).toThrow(grpcError.message);
            expect(t).toThrow(GatewayError);
        });

        it('block events with private data', () => {
            grpcClient.mockBlockEventsWithPrivateDataError(grpcError);

            const t: () => CloseableAsyncIterable<DeliverResponse> = () => gatewayClient.blockEventsWithPrivateData(new Envelope());

            expect(t).toThrow(grpcError.message);
            expect(t).toThrow(GatewayError);
        });
    });
});
