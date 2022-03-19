/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { common, gateway, peer } from '@hyperledger/fabric-protos';
import fs from 'fs';
import os from 'os';
import path from 'path';
import { chaincodeEventsMethod, CloseableAsyncIterable, commitStatusMethod, deliverFilteredMethod, deliverMethod, deliverWithPrivateDataMethod, DuplexStreamResponse, endorseMethod, evaluateMethod, GatewayGrpcClient, ServerStreamResponse, submitMethod } from './client';

/* eslint-disable jest/no-export */

it('Test utilities', () => { // eslint-disable-line jest/expect-expect
    // Empty test to keep Jest happy
});

type MockUnaryRequest<RequestType, ResponseType> = jest.Mock<grpc.ClientUnaryCall, [RequestType, grpc.CallOptions, grpc.requestCallback<ResponseType>]>;
type MockServerStreamRequest<RequestType, ResponseType> = jest.Mock<ServerStreamResponse<ResponseType>, [RequestType, grpc.CallOptions]>;
type MockDuplexStreamRequest<RequestType, ResponseType> = jest.Mock<DuplexStreamResponse<RequestType, ResponseType>, [grpc.CallOptions]>;

const emptyDuplexStreamResponse = {
    async* [Symbol.asyncIterator]() {
        // Nothing
    },
    cancel(): void {
        // Nothing
    },
    write(): boolean {
        return true;
    }
};

const emptyServerStreamResponse = {
    async* [Symbol.asyncIterator]() {
        // Nothing
    },
    cancel(): void {
        // Nothing
    },
};

export class MockGatewayGrpcClient implements GatewayGrpcClient {
    readonly #chaincodeEventsMock = jest.fn() as MockServerStreamRequest<gateway.SignedChaincodeEventsRequest, gateway.ChaincodeEventsResponse>;
    readonly #commitStatusMock = jest.fn() as MockUnaryRequest<gateway.SignedCommitStatusRequest, gateway.CommitStatusResponse>;
    readonly #deliverMock = jest.fn() as MockDuplexStreamRequest<common.Envelope, peer.DeliverResponse>;
    readonly #deliverFilteredMock = jest.fn() as MockDuplexStreamRequest<common.Envelope, peer.DeliverResponse>;
    readonly #deliverWithPrivateDataMock = jest.fn() as MockDuplexStreamRequest<common.Envelope, peer.DeliverResponse>;
    readonly #endorseMock = jest.fn() as MockUnaryRequest<gateway.EndorseRequest, gateway.EndorseResponse>;
    readonly #evaluateMock = jest.fn() as MockUnaryRequest<gateway.EvaluateRequest, gateway.EvaluateResponse>;
    readonly #submitMock = jest.fn() as MockUnaryRequest<gateway.SubmitRequest, gateway.SubmitResponse>;

    #unaryMocks: Record<string, MockUnaryRequest<any, any>> = { // eslint-disable-line @typescript-eslint/no-explicit-any
        [commitStatusMethod]: this.#commitStatusMock,
        [endorseMethod]: this.#endorseMock,
        [evaluateMethod]: this.#evaluateMock,
        [submitMethod]: this.#submitMock,
    };
    #serverStreamMocks: Record<string, MockServerStreamRequest<any, any>> = {  // eslint-disable-line @typescript-eslint/no-explicit-any
        [chaincodeEventsMethod]: this.#chaincodeEventsMock,
    };
    #duplexStreamMocks: Record<string, MockDuplexStreamRequest<any, any>> = {  // eslint-disable-line @typescript-eslint/no-explicit-any
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
        return mock(argument, options, callback);
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
        return mock(argument, options); // eslint-disable-line @typescript-eslint/no-unsafe-return
    }

    makeBidiStreamRequest<RequestType, ResponseType>(
        method: string,
        serialize: (value: RequestType) => Buffer,
        deserialize: (value: Buffer) => ResponseType,
        options: grpc.CallOptions
    ): DuplexStreamResponse<RequestType, ResponseType> {
        const mock = this.#duplexStreamMocks[method];
        if (!mock) {
            throw new Error(`No duplex stream mock for ${method}`);
        }
        return mock(options); // eslint-disable-line @typescript-eslint/no-unsafe-return
    }

    getChaincodeEventsRequests(): gateway.SignedChaincodeEventsRequest[] {
        return this.#chaincodeEventsMock.mock.calls.map(call => call[0]);
    }

    getCommitStatusRequests(): gateway.SignedCommitStatusRequest[] {
        return this.#commitStatusMock.mock.calls.map(call => call[0]);
    }

    getEndorseRequests(): gateway.EndorseRequest[] {
        return this.#endorseMock.mock.calls.map(call => call[0]);
    }

    getEvaluateRequests(): gateway.EvaluateRequest[] {
        return this.#evaluateMock.mock.calls.map(call => call[0]);
    }

    getSubmitRequests(): gateway.SubmitRequest[] {
        return this.#submitMock.mock.calls.map(call => call[0]);
    }

    getBlockEventsOptions(): grpc.CallOptions[] {
        return this.#deliverMock.mock.calls.map(call => call[0]);
    }

    getBlockAndPrivateDataEventsOptions(): grpc.CallOptions[] {
        return this.#deliverWithPrivateDataMock.mock.calls.map(call => call[0]);
    }

    getChaincodeEventsOptions(): grpc.CallOptions[] {
        return this.#chaincodeEventsMock.mock.calls.map(call => call[1]);
    }

    getCommitStatusOptions(): grpc.CallOptions[] {
        return this.#commitStatusMock.mock.calls.map(call => call[1]);
    }

    getEndorseOptions(): grpc.CallOptions[] {
        return this.#endorseMock.mock.calls.map(call => call[1]);
    }

    getEvaluateOptions(): grpc.CallOptions[] {
        return this.#evaluateMock.mock.calls.map(call => call[1]);
    }

    getFilteredBlockEventsOptions(): grpc.CallOptions[] {
        return this.#deliverFilteredMock.mock.calls.map(call => call[0]);
    }

    getSubmitOptions(): grpc.CallOptions[] {
        return this.#submitMock.mock.calls.map(call => call[1]);
    }

    mockCommitStatusResponse(response: gateway.CommitStatusResponse): void {
        this.#commitStatusMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockCommitStatusError(err: grpc.ServiceError): void {
        this.#commitStatusMock.mockImplementation(fakeUnaryCall(err, undefined));
    }

    mockEndorseResponse(response: gateway.EndorseResponse): void {
        this.#endorseMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockEndorseError(err: grpc.ServiceError): void {
        this.#endorseMock.mockImplementation(fakeUnaryCall(err, undefined));
    }

    mockEvaluateResponse(response: gateway.EvaluateResponse): void {
        this.#evaluateMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockEvaluateError(err: grpc.ServiceError): void {
        this.#evaluateMock.mockImplementation(fakeUnaryCall(err, undefined));
    }

    mockSubmitResponse(response: gateway.SubmitResponse): void {
        this.#submitMock.mockImplementation(fakeUnaryCall(undefined, response));
    }

    mockSubmitError(err: grpc.ServiceError): void {
        this.#submitMock.mockImplementation(fakeUnaryCall(err, undefined));
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
        setImmediate(() => callback(err ?? null, response));
        return {} as grpc.ClientUnaryCall;
    };
}

export function newEndorseResponse(options: {
    result: Uint8Array;
    channelName?: string;
}): gateway.EndorseResponse {
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
    close: jest.Mock<void, void[]>;
}

export function newCloseableAsyncIterable<T>(values: T[]): CloseableAsyncIterableStub<T> {
    return Object.assign(newAsyncIterable(values), {
        close: jest.fn<void, void[]>(),
    });
}

export interface ServerStreamResponseStub<T> extends ServerStreamResponse<T> {
    cancel: jest.Mock<void, void[]>;
}

export function newServerStreamResponse<T>(values: (T | Error)[]): ServerStreamResponseStub<T> {
    return Object.assign(newAsyncIterable(values), {
        cancel: jest.fn<void, void[]>(),
    });
}

function newAsyncIterable<T>(values: (T | Error)[]): AsyncIterable<T> {
    return {
        async* [Symbol.asyncIterator]() { // eslint-disable-line @typescript-eslint/require-await
            for (const value of values) {
                if (value instanceof Error) {
                    throw value;
                }
                yield value;
            }
        }
    };
}

export interface DuplexStreamResponseStub<RequestType, ResponseType> extends DuplexStreamResponse<RequestType, ResponseType> {
    cancel: jest.Mock<void, void[]>;
    write: jest.Mock<boolean, RequestType[]>;
}

export function newDuplexStreamResponse<RequestType, ResponseType>(values: (ResponseType | Error)[]): DuplexStreamResponseStub<RequestType, ResponseType> {
    return Object.assign(newServerStreamResponse(values), {
        write: jest.fn<boolean, RequestType[]>(),
    });
}

export async function createTempDir(): Promise<string> {
    const prefix = `${os.tmpdir()}${path.sep}`;
    return await fs.promises.mkdtemp(prefix);
}
