/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable jest/no-export */

import * as grpc from '@grpc/grpc-js';
import { UnaryCallback } from '@grpc/grpc-js/build/src/client';
import { chaincodeEventsMethod, commitStatusMethod, endorseMethod, evaluateMethod, GatewayClient, GatewayGrpcClient, newGatewayClient, ServerStreamResponse, submitMethod } from './client';
import { GatewayError } from './gatewayerror';
import { ChaincodeEventsResponse, CommitStatusResponse, EndorseRequest, EndorseResponse, EvaluateRequest, EvaluateResponse, SignedChaincodeEventsRequest, SignedCommitStatusRequest, SubmitRequest, SubmitResponse } from './protos/gateway/gateway_pb';

type MockUnaryRequest<RequestType, ResponseType> = jest.Mock<grpc.ClientUnaryCall, [RequestType, grpc.CallOptions, UnaryCallback<ResponseType>]>;
type MockServerStreamRequest<RequestType, ResponseType> = jest.Mock<ServerStreamResponse<ResponseType>, [RequestType, grpc.CallOptions]>;

export class MockGatewayGrpcClient implements GatewayGrpcClient {
    readonly chaincodeEventsMock = jest.fn() as MockServerStreamRequest<SignedChaincodeEventsRequest, ChaincodeEventsResponse>;
    readonly commitStatusMock = jest.fn() as MockUnaryRequest<SignedCommitStatusRequest, CommitStatusResponse>;
    readonly endorseMock = jest.fn() as MockUnaryRequest<EndorseRequest, EndorseResponse>;
    readonly evaluateMock = jest.fn() as MockUnaryRequest<EvaluateRequest, EvaluateResponse>;
    readonly submitMock = jest.fn() as MockUnaryRequest<SubmitRequest, SubmitResponse>;

    #unaryMocks = {
        [commitStatusMethod]: this.commitStatusMock,
        [endorseMethod]: this.endorseMock,
        [evaluateMethod]: this.evaluateMock,
        [submitMethod]: this.submitMock,
    };
    #serverStreamMocks = {
        [chaincodeEventsMethod]: this.chaincodeEventsMock,
    };

    constructor() {
        // Default empty responses
        this.chaincodeEventsMock.mockReturnValue({
            async* [Symbol.asyncIterator]() {
                // Nothing
            },
            cancel(): void {
                // Nothing
            },
        });
        this.mockCommitStatusResponse(new CommitStatusResponse());
        this.mockEndorseResponse(new EndorseResponse());
        this.mockEvaluateResponse(new EvaluateResponse());
        this.mockSubmitResponse(new SubmitResponse());
    }

    makeUnaryRequest<RequestType, ResponseType>(
        method: string,
        serialize: (value: RequestType) => Buffer,
        deserialize: (value: Buffer) => ResponseType,
        argument: RequestType,
        options: grpc.CallOptions,
        callback: grpc.requestCallback<ResponseType>
    ): grpc.ClientUnaryCall {
        const mock = this.#unaryMocks[method];
        if (!mock) {
            throw new Error(`No unary mock for ${method}`);
        }
        // eslint-disable-next-line @typescript-eslint/no-unsafe-argument,@typescript-eslint/no-explicit-any
        return mock(argument as any, options, callback as any);
    }

    makeServerStreamRequest<RequestType, ResponseType>(
        method: string,
        serialize: (value: RequestType) => Buffer,
        deserialize: (value: Buffer) => ResponseType,
        argument: RequestType,
        options: grpc.CallOptions
    ): ServerStreamResponse<ResponseType> {
        const mock = this.#serverStreamMocks[method];
        if (!mock) {
            throw new Error(`No server stream mock for ${method}`);
        }
        // eslint-disable-next-line @typescript-eslint/no-unsafe-return,@typescript-eslint/no-unsafe-argument,@typescript-eslint/no-explicit-any
        return mock(argument as any, options) as any;
    }

    getChaincodeEventsRequests(): SignedChaincodeEventsRequest[] {
        return this.chaincodeEventsMock.mock.calls.map(call => call[0]);
    }

    getCommitStatusRequests(): SignedCommitStatusRequest[] {
        return this.commitStatusMock.mock.calls.map(call => call[0]);
    }

    getEndorseRequests(): EndorseRequest[] {
        return this.endorseMock.mock.calls.map(call => call[0]);
    }

    getEvaluateRequests(): EvaluateRequest[] {
        return this.evaluateMock.mock.calls.map(call => call[0]);
    }

    getSubmitRequests(): SubmitRequest[] {
        return this.submitMock.mock.calls.map(call => call[0]);
    }

    getChaincodeEventsOptions(): grpc.CallOptions[] {
        return this.chaincodeEventsMock.mock.calls.map(call => call[1]);
    }

    getCommitStatusOptions(): grpc.CallOptions[] {
        return this.commitStatusMock.mock.calls.map(call => call[1]);
    }

    getEndorseOptions(): grpc.CallOptions[] {
        return this.endorseMock.mock.calls.map(call => call[1]);
    }

    getEvaluateOptions(): grpc.CallOptions[] {
        return this.evaluateMock.mock.calls.map(call => call[1]);
    }

    getSubmitOptions(): grpc.CallOptions[] {
        return this.submitMock.mock.calls.map(call => call[1]);
    }

    mockCommitStatusResponse(response: CommitStatusResponse): void {
        this.commitStatusMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockCommitStatusError(err: grpc.ServiceError): void {
        this.commitStatusMock.mockImplementation(fakeUnaryCall(err, undefined));
    }

    mockEndorseResponse(response: EndorseResponse): void {
        this.endorseMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockEndorseError(err: grpc.ServiceError): void {
        this.endorseMock.mockImplementation(fakeUnaryCall(err, undefined));
    }

    mockEvaluateResponse(response: EvaluateResponse): void {
        this.evaluateMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockEvaluateError(err: grpc.ServiceError): void {
        this.evaluateMock.mockImplementation(fakeUnaryCall(err, undefined));
    }

    mockSubmitResponse(response: SubmitResponse): void {
        this.submitMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockSubmitError(err: grpc.ServiceError): void {
        this.submitMock.mockImplementation(fakeUnaryCall(err, undefined));
    }

    mockChaincodeEventsResponse(stream: ServerStreamResponse<ChaincodeEventsResponse>): void {
        this.chaincodeEventsMock.mockReturnValue(stream);
    }
}

function fakeUnaryCall<ResponseType>(err: grpc.ServiceError | undefined, response: ResponseType | undefined) {
    return (request: unknown, options: grpc.CallOptions, callback: UnaryCallback<ResponseType>) => {
        setImmediate(() => callback(err || null, response))
        return {} as grpc.ClientUnaryCall;
    };
}

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

            const t = () => gatewayClient.evaluate(new EvaluateRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('endorse', async () => {
            grpcClient.mockEndorseError(grpcError);

            const t = () => gatewayClient.endorse(new EndorseRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('submit', async () => {
            grpcClient.mockSubmitError(grpcError);

            const t = () => gatewayClient.submit(new SubmitRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });

        it('commit status', async () => {
            grpcClient.mockCommitStatusError(grpcError);

            const t = () => gatewayClient.commitStatus(new SignedCommitStatusRequest());

            await expect(t).rejects.toThrow(grpcError.message);
            await expect(t).rejects.toThrow(GatewayError);
        });
    });
});
