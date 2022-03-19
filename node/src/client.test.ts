/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable jest/no-export */

import * as grpc from '@grpc/grpc-js';
import { common, gateway } from '@hyperledger/fabric-protos';
import { GatewayClient, newGatewayClient } from './client';
import { GatewayError } from './gatewayerror';
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

            const t = gatewayClient.evaluate(new gateway.EvaluateRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('endorse', async () => {
            grpcClient.mockEndorseError(grpcError);

            const t = gatewayClient.endorse(new gateway.EndorseRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('submit', async () => {
            grpcClient.mockSubmitError(grpcError);

            const t = gatewayClient.submit(new gateway.SubmitRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('commit status', async () => {
            grpcClient.mockCommitStatusError(grpcError);

            const t = gatewayClient.commitStatus(new gateway.SignedCommitStatusRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('chaincode events', () => {
            grpcClient.mockChaincodeEventsError(grpcError);

            const t: () => void = () => gatewayClient.chaincodeEvents(new gateway.SignedChaincodeEventsRequest());

            expect(t).toThrow(grpcError.message);
            expect(t).toThrow(GatewayError);
        });

        it('block events', () => {
            grpcClient.mockBlockEventsError(grpcError);

            const t: () => void = () => gatewayClient.blockEvents(new common.Envelope());

            expect(t).toThrow(grpcError.message);
            expect(t).toThrow(GatewayError);
        });

        it('filtered block events', () => {
            grpcClient.mockFilteredBlockEventsError(grpcError);

            const t: () => void = () => gatewayClient.filteredBlockEvents(new common.Envelope());

            expect(t).toThrow(grpcError.message);
            expect(t).toThrow(GatewayError);
        });

        it('block and private data events', () => {
            grpcClient.mockBlockAndPrivateDataEventsError(grpcError);

            const t: () => void = () => gatewayClient.blockAndPrivateDataEvents(new common.Envelope());

            expect(t).toThrow(grpcError.message);
            expect(t).toThrow(GatewayError);
        });
    });
});
