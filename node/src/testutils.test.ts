/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { common, gateway, peer } from '@hyperledger/fabric-protos';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {
    CloseableAsyncIterable,
    DuplexStreamResponse,
    GatewayGrpcClient,
    ServerStreamResponse,
    chaincodeEventsMethod,
    commitStatusMethod,
    deliverFilteredMethod,
    deliverMethod,
    deliverWithPrivateDataMethod,
    endorseMethod,
    evaluateMethod,
    submitMethod,
} from './client';
import { assertDefined } from './gateway';

/* eslint-disable jest/no-export */

// eslint-disable-next-line jest/expect-expect
it('Test utilities', () => {
    // Empty test to keep Jest happy
});

type MockUnaryRequest<RequestType, ResponseType> = jest.Mock<
    grpc.ClientUnaryCall,
    [RequestType, grpc.CallOptions, grpc.requestCallback<ResponseType>]
>;
type MockServerStreamRequest<RequestType, ResponseType> = jest.Mock<
    ServerStreamResponse<ResponseType>,
    [RequestType, grpc.CallOptions]
>;
type MockDuplexStreamRequest<RequestType, ResponseType> = jest.Mock<
    DuplexStreamResponse<RequestType, ResponseType>,
    [grpc.CallOptions]
>;

const emptyDuplexStreamResponse = {
    async *[Symbol.asyncIterator]() {
        // Nothing
    },
    cancel(): void {
        // Nothing
    },
    write(): boolean {
        return true;
    },
};

const emptyServerStreamResponse = {
    async *[Symbol.asyncIterator]() {
        // Nothing
    },
    cancel(): void {
        // Nothing
    },
};

export class MockGatewayGrpcClient implements GatewayGrpcClient {
    readonly #chaincodeEventsMock = jest.fn() as MockServerStreamRequest<
        gateway.SignedChaincodeEventsRequest,
        gateway.ChaincodeEventsResponse
    >;
    readonly #commitStatusMock = jest.fn() as MockUnaryRequest<
        gateway.SignedCommitStatusRequest,
        gateway.CommitStatusResponse
    >;
    readonly #deliverMock = jest.fn() as MockDuplexStreamRequest<common.Envelope, peer.DeliverResponse>;
    readonly #deliverFilteredMock = jest.fn() as MockDuplexStreamRequest<common.Envelope, peer.DeliverResponse>;
    readonly #deliverWithPrivateDataMock = jest.fn() as MockDuplexStreamRequest<common.Envelope, peer.DeliverResponse>;
    readonly #endorseMock = jest.fn() as MockUnaryRequest<gateway.EndorseRequest, gateway.EndorseResponse>;
    readonly #evaluateMock = jest.fn() as MockUnaryRequest<gateway.EvaluateRequest, gateway.EvaluateResponse>;
    readonly #submitMock = jest.fn() as MockUnaryRequest<gateway.SubmitRequest, gateway.SubmitResponse>;

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    #unaryMocks: Record<string, MockUnaryRequest<any, any>> = {
        [commitStatusMethod]: this.#commitStatusMock,
        [endorseMethod]: this.#endorseMock,
        [evaluateMethod]: this.#evaluateMock,
        [submitMethod]: this.#submitMock,
    };
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    #serverStreamMocks: Record<string, MockServerStreamRequest<any, any>> = {
        [chaincodeEventsMethod]: this.#chaincodeEventsMock,
    };
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    #duplexStreamMocks: Record<string, MockDuplexStreamRequest<any, any>> = {
        [deliverMethod]: this.#deliverMock,
        [deliverFilteredMethod]: this.#deliverFilteredMock,
        [deliverWithPrivateDataMethod]: this.#deliverWithPrivateDataMock,
    };

    constructor() {
        // Default empty responses
        this.mockBlockEventsResponse(emptyDuplexStreamResponse);
        this.mockBlockAndPrivateDataEventsResponse(emptyDuplexStreamResponse);
        this.mockChaincodeEventsResponse(emptyServerStreamResponse);
        this.mockCommitStatusResponse(new gateway.CommitStatusResponse());
        this.mockEndorseResponse(new gateway.EndorseResponse());
        this.mockEvaluateResponse(new gateway.EvaluateResponse());
        this.mockFilteredBlockEventsResponse(emptyDuplexStreamResponse);
        this.mockSubmitResponse(new gateway.SubmitResponse());
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    #getUnaryMock(method: string): MockUnaryRequest<any, any> {
        return assertDefined(this.#unaryMocks[method], `no mock defined for method: ${method}`);
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    #getServerStreamMock(method: string): MockServerStreamRequest<any, any> {
        return assertDefined(this.#serverStreamMocks[method], `no mock defined for method: ${method}`);
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    #getDuplexStreamMock(method: string): MockDuplexStreamRequest<any, any> {
        return assertDefined(this.#duplexStreamMocks[method], `no mock defined for method: ${method}`);
    }

    makeUnaryRequest<RequestType, ResponseType>(
        method: string,
        serialize: (value: RequestType) => Buffer,
        deserialize: (value: Buffer) => ResponseType,
        argument: RequestType,
        options: grpc.CallOptions,
        callback: grpc.requestCallback<ResponseType>,
    ): grpc.ClientUnaryCall {
        const mock = this.#getUnaryMock(method);
        return mock(argument, options, callback);
    }

    makeServerStreamRequest<RequestType, ResponseType>(
        method: string,
        serialize: (value: RequestType) => Buffer,
        deserialize: (value: Buffer) => ResponseType,
        argument: RequestType,
        options: grpc.CallOptions,
    ): ServerStreamResponse<ResponseType> {
        const mock = this.#getServerStreamMock(method);
        return mock(argument, options); // eslint-disable-line @typescript-eslint/no-unsafe-return
    }

    makeBidiStreamRequest<RequestType, ResponseType>(
        method: string,
        serialize: (value: RequestType) => Buffer,
        deserialize: (value: Buffer) => ResponseType,
        options: grpc.CallOptions,
    ): DuplexStreamResponse<RequestType, ResponseType> {
        const mock = this.#getDuplexStreamMock(method);
        return mock(options); // eslint-disable-line @typescript-eslint/no-unsafe-return
    }

    getChaincodeEventsRequests(): gateway.SignedChaincodeEventsRequest[] {
        return this.#chaincodeEventsMock.mock.calls.map((call) => call[0]);
    }

    getChaincodeEventsRequest(index = 0): gateway.SignedChaincodeEventsRequest {
        const result = this.getChaincodeEventsRequests()[index];
        return assertDefined(result, `no chaincode events request at index ${String(index)}`);
    }

    getCommitStatusRequests(): gateway.SignedCommitStatusRequest[] {
        return this.#commitStatusMock.mock.calls.map((call) => call[0]);
    }

    getCommitStatusRequest(index = 0): gateway.SignedCommitStatusRequest {
        const result = this.getCommitStatusRequests()[index];
        return assertDefined(result, `no commit status request at index ${String(index)}`);
    }

    getEndorseRequests(): gateway.EndorseRequest[] {
        return this.#endorseMock.mock.calls.map((call) => call[0]);
    }

    getEndorseRequest(index = 0): gateway.EndorseRequest {
        const result = this.getEndorseRequests()[index];
        return assertDefined(result, `no endorse request at index ${String(index)}`);
    }

    getEvaluateRequests(): gateway.EvaluateRequest[] {
        return this.#evaluateMock.mock.calls.map((call) => call[0]);
    }

    getEvaluateRequest(index = 0): gateway.EvaluateRequest {
        const result = this.getEvaluateRequests()[index];
        return assertDefined(result, `no evaluate request at index ${String(index)}`);
    }

    getSubmitRequests(): gateway.SubmitRequest[] {
        return this.#submitMock.mock.calls.map((call) => call[0]);
    }

    getSubmitRequest(index = 0): gateway.SubmitRequest {
        const result = this.getSubmitRequests()[index];
        return assertDefined(result, `no submit request at index ${String(index)}`);
    }

    getBlockEventsOptions(): grpc.CallOptions[] {
        return this.#deliverMock.mock.calls.map((call) => call[0]);
    }

    getBlockEventsOption(index = 0): grpc.CallOptions {
        const result = this.getBlockEventsOptions()[index];
        return assertDefined(result, `no block events option at index ${String(index)}`);
    }

    getBlockAndPrivateDataEventsOptions(): grpc.CallOptions[] {
        return this.#deliverWithPrivateDataMock.mock.calls.map((call) => call[0]);
    }

    getBlockAndPrivateDataEventsOption(index = 0): grpc.CallOptions {
        const result = this.getBlockAndPrivateDataEventsOptions()[index];
        return assertDefined(result, `no block and private data events option at index ${String(index)}`);
    }

    getChaincodeEventsOptions(): grpc.CallOptions[] {
        return this.#chaincodeEventsMock.mock.calls.map((call) => call[1]);
    }

    getChaincodeEventsOption(index = 0): grpc.CallOptions {
        const result = this.getChaincodeEventsOptions()[index];
        return assertDefined(result, `no chaincode events option at index ${String(index)}`);
    }

    getCommitStatusOptions(): grpc.CallOptions[] {
        return this.#commitStatusMock.mock.calls.map((call) => call[1]);
    }

    getCommitStatusOption(index = 0): grpc.CallOptions {
        const result = this.getCommitStatusOptions()[index];
        return assertDefined(result, `no commit status option at index ${String(index)}`);
    }

    getEndorseOptions(): grpc.CallOptions[] {
        return this.#endorseMock.mock.calls.map((call) => call[1]);
    }

    getEndorseOption(index = 0): grpc.CallOptions {
        const result = this.getEndorseOptions()[index];
        return assertDefined(result, `no endorse option at index ${String(index)}`);
    }

    getEvaluateOptions(): grpc.CallOptions[] {
        return this.#evaluateMock.mock.calls.map((call) => call[1]);
    }

    getEvaluateOption(index = 0): grpc.CallOptions {
        const result = this.getEvaluateOptions()[index];
        return assertDefined(result, `no evaluate option at index ${String(index)}`);
    }

    getFilteredBlockEventsOptions(): grpc.CallOptions[] {
        return this.#deliverFilteredMock.mock.calls.map((call) => call[0]);
    }

    getFilteredBlockEventsOption(index = 0): grpc.CallOptions {
        const result = this.getFilteredBlockEventsOptions()[index];
        return assertDefined(result, `no filtered block events option at index ${String(index)}`);
    }

    getSubmitOptions(): grpc.CallOptions[] {
        return this.#submitMock.mock.calls.map((call) => call[1]);
    }

    getSubmitOption(index = 0): grpc.CallOptions {
        const result = this.getSubmitOptions()[index];
        return assertDefined(result, `no submit option at index ${String(index)}`);
    }

    mockCommitStatusResponse(response: gateway.CommitStatusResponse): void {
        this.#commitStatusMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockCommitStatusError(err: grpc.ServiceError): void {
        this.#commitStatusMock.mockImplementation(fakeUnaryCall<gateway.CommitStatusResponse>(err, undefined));
    }

    mockEndorseResponse(response: gateway.EndorseResponse): void {
        this.#endorseMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockEndorseError(err: grpc.ServiceError): void {
        this.#endorseMock.mockImplementation(fakeUnaryCall<gateway.EndorseResponse>(err, undefined));
    }

    mockEvaluateResponse(response: gateway.EvaluateResponse): void {
        this.#evaluateMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockEvaluateError(err: grpc.ServiceError): void {
        this.#evaluateMock.mockImplementation(fakeUnaryCall<gateway.EvaluateResponse>(err, undefined));
    }

    mockSubmitResponse(response: gateway.SubmitResponse): void {
        this.#submitMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockSubmitError(err: grpc.ServiceError): void {
        this.#submitMock.mockImplementation(fakeUnaryCall<gateway.SubmitResponse>(err, undefined));
    }

    mockChaincodeEventsResponse(stream: ServerStreamResponse<gateway.ChaincodeEventsResponse>): void {
        this.#chaincodeEventsMock.mockReturnValue(stream);
    }

    mockChaincodeEventsError(err: grpc.ServiceError): void {
        this.#chaincodeEventsMock.mockImplementation(() => {
            throw err;
        });
    }

    mockBlockEventsResponse(stream: DuplexStreamResponse<common.Envelope, peer.DeliverResponse>): void {
        this.#deliverMock.mockReturnValue(stream);
    }

    mockBlockEventsError(err: grpc.ServiceError): void {
        this.#deliverMock.mockImplementation(() => {
            throw err;
        });
    }

    mockFilteredBlockEventsResponse(stream: DuplexStreamResponse<common.Envelope, peer.DeliverResponse>): void {
        this.#deliverFilteredMock.mockReturnValue(stream);
    }

    mockFilteredBlockEventsError(err: grpc.ServiceError): void {
        this.#deliverFilteredMock.mockImplementation(() => {
            throw err;
        });
    }

    mockBlockAndPrivateDataEventsResponse(stream: DuplexStreamResponse<common.Envelope, peer.DeliverResponse>): void {
        this.#deliverWithPrivateDataMock.mockReturnValue(stream);
    }

    mockBlockAndPrivateDataEventsError(err: grpc.ServiceError): void {
        this.#deliverWithPrivateDataMock.mockImplementation(() => {
            throw err;
        });
    }
}

function fakeUnaryCall<ResponseType>(err: grpc.ServiceError | undefined, response: ResponseType | undefined) {
    return (request: unknown, options: grpc.CallOptions, callback: grpc.requestCallback<ResponseType>) => {
        setImmediate(() => {
            callback(err ?? null, response);
        });
        return {} as grpc.ClientUnaryCall;
    };
}

export function newEndorseResponse(options: { result: Uint8Array; channelName?: string }): gateway.EndorseResponse {
    const chaincodeResponse = new peer.Response();
    chaincodeResponse.setPayload(options.result);

    const chaincodeAction = new peer.ChaincodeAction();
    chaincodeAction.setResponse(chaincodeResponse);

    const responsePayload = new peer.ProposalResponsePayload();
    responsePayload.setExtension$(chaincodeAction.serializeBinary());

    const endorsedAction = new peer.ChaincodeEndorsedAction();
    endorsedAction.setProposalResponsePayload(responsePayload.serializeBinary());

    const actionPayload = new peer.ChaincodeActionPayload();
    actionPayload.setAction(endorsedAction);

    const transactionAction = new peer.TransactionAction();
    transactionAction.setPayload(actionPayload.serializeBinary());

    const transaction = new peer.Transaction();
    transaction.setActionsList([transactionAction]);

    const payload = new common.Payload();
    payload.setData(transaction.serializeBinary());

    const channelHeader = new common.ChannelHeader();
    channelHeader.setChannelId(options.channelName ?? 'network');

    const header = new common.Header();
    header.setChannelHeader(channelHeader.serializeBinary());

    payload.setHeader(header);

    const envelope = new common.Envelope();
    envelope.setPayload(payload.serializeBinary());

    const endorseResponse = new gateway.EndorseResponse();
    endorseResponse.setPreparedTransaction(envelope);

    return endorseResponse;
}

export async function readElements<T>(
    iter: AsyncIterable<T>,
    count: number,
    onRead?: (element: T) => Promise<void>,
): Promise<T[]> {
    const elements: T[] = [];
    for await (const element of iter) {
        elements.push(element);
        await onRead?.(element);

        if (--count <= 0) {
            break;
        }
    }

    return elements;
}

export interface CloseableAsyncIterableStub<T> extends CloseableAsyncIterable<T> {
    close: jest.Mock<void, []>;
}

// @ts-expect-error Polyfill for Symbol.dispose if not present
Symbol.dispose ??= Symbol('Symbol.dispose');

export function newCloseableAsyncIterable<T>(values: T[]): CloseableAsyncIterableStub<T> {
    return Object.assign(newAsyncIterable(values), {
        close: jest.fn<undefined, []>(),
        [Symbol.dispose]: jest.fn<undefined, []>(),
    });
}

export interface ServerStreamResponseStub<T> extends ServerStreamResponse<T> {
    cancel: jest.Mock<undefined, []>;
}

export function newServerStreamResponse<T>(values: (T | Error)[]): ServerStreamResponseStub<T> {
    return Object.assign(newAsyncIterable(values), {
        cancel: jest.fn<undefined, []>(),
    });
}

function newAsyncIterable<T>(values: (T | Error)[]): AsyncIterable<T> {
    return {
        [Symbol.asyncIterator]: async function* () {
            for (const value of values) {
                if (value instanceof Error) {
                    return Promise.reject(value);
                }
                yield value;
            }
        },
    };
}

export interface DuplexStreamResponseStub<RequestType, ResponseType>
    extends DuplexStreamResponse<RequestType, ResponseType> {
    cancel: jest.Mock<undefined, []>;
    write: jest.Mock<boolean, RequestType[]>;
    getRequest(index?: number): RequestType;
}

export function newDuplexStreamResponse<RequestType, ResponseType>(
    values: (ResponseType | Error)[],
): DuplexStreamResponseStub<RequestType, ResponseType> {
    const write = jest.fn<boolean, RequestType[]>();

    return Object.assign(newServerStreamResponse(values), {
        write,
        getRequest: (index = 0) => {
            const call = write.mock.calls[index];
            const result = assertDefined(call, `no request at index ${String(index)}`)[0];
            return assertDefined(result, `no argument for request at index ${String(index)}`);
        },
    });
}

export async function createTempDir(): Promise<string> {
    const prefix = `${os.tmpdir()}${path.sep}`;
    return await fs.promises.mkdtemp(prefix);
}
