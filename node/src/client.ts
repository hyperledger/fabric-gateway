/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions, ClientUnaryCall, Metadata, requestCallback, ServiceError } from '@grpc/grpc-js';
import { Message } from 'google-protobuf';
import { CommitStatusError } from './commitstatuserror';
import { EndorseError } from './endorseerror';
import { ConnectOptions } from './gateway';
import { GatewayError, newGatewayError } from './gatewayerror';
import { Envelope } from './protos/common/common_pb';
import { ChaincodeEventsResponse, CommitStatusRequest, CommitStatusResponse, EndorseRequest, EndorseResponse, EvaluateRequest, EvaluateResponse, SignedChaincodeEventsRequest, SignedCommitStatusRequest, SubmitRequest, SubmitResponse } from './protos/gateway/gateway_pb';
import { DeliverResponse } from './protos/peer/events_pb';
import { SubmitError } from './submiterror';

export const evaluateMethod = '/gateway.Gateway/Evaluate';
export const endorseMethod = '/gateway.Gateway/Endorse';
export const submitMethod = '/gateway.Gateway/Submit';
export const commitStatusMethod = '/gateway.Gateway/CommitStatus';
export const chaincodeEventsMethod = '/gateway.Gateway/ChaincodeEvents';
export const deliverMethod = '/protos.Deliver/Deliver';
export const deliverFilteredMethod = '/protos.Deliver/DeliverFiltered';
export const deliverWithPrivateDataMethod = '/protos.Deliver/DeliverWithPrivateData';

export interface GatewayClient {
    evaluate(request: EvaluateRequest, options?: CallOptions): Promise<EvaluateResponse>;
    endorse(request: EndorseRequest, options?: CallOptions): Promise<EndorseResponse>;
    submit(request: SubmitRequest, options?: CallOptions): Promise<SubmitResponse>;
    commitStatus(request: SignedCommitStatusRequest, options?: CallOptions): Promise<CommitStatusResponse>;
    chaincodeEvents(request: SignedChaincodeEventsRequest, options?: CallOptions): CloseableAsyncIterable<ChaincodeEventsResponse>;
    blockEvents(request: Envelope, options?: CallOptions): CloseableAsyncIterable<DeliverResponse>;
    filteredBlockEvents(request: Envelope, options?: CallOptions): CloseableAsyncIterable<DeliverResponse>;
    blockEventsWithPrivateData(request: Envelope, options?: CallOptions): CloseableAsyncIterable<DeliverResponse>;
}

/**
 * An async iterable that can be closed when the consumer does not want to read any more elements, freeing up resources
 * that may be held by the iterable.
 */
export interface CloseableAsyncIterable<T> extends AsyncIterable<T> {
    /**
     * Close the iterable to free up resources when no more elements are required.
     */
    close(): void;
}

/**
 * Subset of grpc-js ClientReadableStream used by GatewayClient to aid mocking.
 */
export interface ServerStreamResponse<T> extends AsyncIterable<T> {
    cancel(): void;
}

/**
 * Subset of grpc-js ClientDuplexStream used by GatewayClient to aid mocking.
 */
export interface DuplexStreamResponse<RequestType, ResponseType> extends ServerStreamResponse<ResponseType> {
    write(chunk: RequestType): boolean;
}

/**
 * Subset of the grpc-js Client used by GatewayClient to aid mocking.
 */
export interface GatewayGrpcClient {
    makeUnaryRequest<RequestType, ResponseType>(method: string, serialize: (value: RequestType) => Buffer, deserialize: (value: Buffer) => ResponseType, argument: RequestType, options: CallOptions, callback: requestCallback<ResponseType>): ClientUnaryCall;
    makeServerStreamRequest<RequestType, ResponseType>(method: string, serialize: (value: RequestType) => Buffer, deserialize: (value: Buffer) => ResponseType, argument: RequestType, options: CallOptions): ServerStreamResponse<ResponseType>;
    makeBidiStreamRequest<RequestType, ResponseType>(method: string, serialize: (value: RequestType) => Buffer, deserialize: (value: Buffer) => ResponseType, options: CallOptions): DuplexStreamResponse<RequestType, ResponseType>;
}

type DefaultCallOptions = Pick<ConnectOptions, 'commitStatusOptions' | 'endorseOptions' | 'evaluateOptions' | 'submitOptions' | 'chaincodeEventsOptions' | 'blockEventsOptions' | 'filteredBlockEventsOptions' | 'blockEventsWithPrivateDataOptions'>;

class GatewayClientImpl implements GatewayClient {
    readonly #client: GatewayGrpcClient;
    readonly #defaultOptions: Readonly<DefaultCallOptions>;

    constructor(client: GatewayGrpcClient, defaultOptions: Readonly<DefaultCallOptions>) {
        this.#client = client;
        this.#defaultOptions = Object.assign({}, defaultOptions);
    }

    evaluate(request: EvaluateRequest, options?: Readonly<CallOptions>): Promise<EvaluateResponse> {
        return new Promise((resolve, reject) => this.#client.makeUnaryRequest(
            evaluateMethod,
            serialize,
            deserializeEvaluateResponse,
            request,
            buildOptions(this.#defaultOptions.evaluateOptions, options),
            newUnaryCallback(resolve, reject)
        ));
    }

    endorse(request: EndorseRequest, options?: Readonly<CallOptions>): Promise<EndorseResponse> {
        return new Promise((resolve, reject) => this.#client.makeUnaryRequest(
            endorseMethod,
            serialize,
            deserializeEndorseResponse,
            request,
            buildOptions(this.#defaultOptions.endorseOptions, options),
            newUnaryCallback(
                resolve,
                reject,
                err => new EndorseError(Object.assign(err, { transactionId: request.getTransactionId() }))
            ),
        ));
    }

    submit(request: SubmitRequest, options?: Readonly<CallOptions>): Promise<SubmitResponse> {
        return new Promise((resolve, reject) => this.#client.makeUnaryRequest(
            submitMethod,
            serialize,
            deserializeSubmitResponse,
            request,
            buildOptions(this.#defaultOptions.submitOptions, options),
            newUnaryCallback(
                resolve,
                reject,
                err => new SubmitError(Object.assign(err, { transactionId: request.getTransactionId() })),
            ),
        ));
    }

    commitStatus(request: SignedCommitStatusRequest, options?: Readonly<CallOptions>): Promise<CommitStatusResponse> {
        return new Promise((resolve, reject) => this.#client.makeUnaryRequest(
            commitStatusMethod,
            serialize,
            deserializeCommitStatusResponse,
            request,
            buildOptions(this.#defaultOptions.commitStatusOptions, options),
            newUnaryCallback(
                resolve,
                reject,
                err => {
                    const req = CommitStatusRequest.deserializeBinary(request.getRequest_asU8());
                    return new CommitStatusError(Object.assign(err, { transactionId: req.getTransactionId() }));
                },
            )
        ));
    }

    chaincodeEvents(request: SignedChaincodeEventsRequest, options?: Readonly<CallOptions>): CloseableAsyncIterable<ChaincodeEventsResponse> {
        return this.#makeServerStreamRequest(
            chaincodeEventsMethod,
            deserializeChaincodeEventsResponse,
            request,
            buildOptions(this.#defaultOptions.chaincodeEventsOptions, options),
        );
    }

    #makeServerStreamRequest<RequestType extends Message, ResponseType>(
        method: string,
        deserialize: (value: Buffer) => ResponseType,
        argument: RequestType,
        options: CallOptions
    ): CloseableAsyncIterable<ResponseType> {
        try {
            const serverStream = this.#client.makeServerStreamRequest(method, serialize, deserialize, argument, options);
            return {
                [Symbol.asyncIterator]: () => wrapAsyncIterator(serverStream[Symbol.asyncIterator]()),
                close: () => serverStream.cancel(),
            };
        } catch (err) {
            rethrowGrpcError(err);
        }
    }

    blockEvents(request: Envelope, options?: CallOptions): CloseableAsyncIterable<DeliverResponse> {
        return this.#makeBidiStreamRequest(
            deliverMethod,
            deserializeDeliverResponse,
            request,
            buildOptions(this.#defaultOptions.blockEventsOptions, options),
        );
    }

    filteredBlockEvents(request: Envelope, options?: CallOptions): CloseableAsyncIterable<DeliverResponse> {
        return this.#makeBidiStreamRequest(
            deliverFilteredMethod,
            deserializeDeliverResponse,
            request,
            buildOptions(this.#defaultOptions.filteredBlockEventsOptions, options),
        );
    }

    blockEventsWithPrivateData(request: Envelope, options?: CallOptions): CloseableAsyncIterable<DeliverResponse> {
        return this.#makeBidiStreamRequest(
            deliverWithPrivateDataMethod,
            deserializeDeliverResponse,
            request,
            buildOptions(this.#defaultOptions.blockEventsWithPrivateDataOptions, options),
        );
    }

    #makeBidiStreamRequest<RequestType extends Message, ResponseType>(
        method: string,
        deserialize: (value: Buffer) => ResponseType,
        request: RequestType,
        options: CallOptions
    ): CloseableAsyncIterable<ResponseType> {
        try {
            const duplexStream = this.#client.makeBidiStreamRequest(method, serialize, deserialize, options);
            duplexStream.write(request);
            return {
                [Symbol.asyncIterator]: () => wrapAsyncIterator(duplexStream[Symbol.asyncIterator]()),
                close: () => duplexStream.cancel(),
            };
        } catch (err) {
            rethrowGrpcError(err);
        }
    }
}

function buildOptions(defaultOptions: (() => CallOptions) | undefined, options?: Readonly<CallOptions>): CallOptions {
    return Object.assign({}, defaultOptions?.(), options);
}

function newUnaryCallback<T>(
    resolve: (value: T) => void,
    reject: (reason: Error) => void,
    wrap: (err: GatewayError) => GatewayError = (err => err)
): requestCallback<T> {
    return (err, value) => {
        if (err) {
            return reject(wrap(newGatewayError(err)));
        }
        if (value == null) {
            return reject(new Error('No result returned'));
        }
        return resolve(value);
    };
}

function wrapAsyncIterator<T>(iterator: AsyncIterator<T>): AsyncIterator<T> {
    return {
        next: async (...args) => {
            try {
                return await iterator.next(...args);
            } catch (err) {
                rethrowGrpcError(err);
            }
        }
    };
}

function rethrowGrpcError(err: unknown): never {
    if (isServiceError(err)) {
        throw newGatewayError(err);
    }
    throw err;
}

function isServiceError(err: unknown): err is ServiceError {
    return typeof (err as ServiceError).code === 'number' &&
        typeof (err as ServiceError).details === 'string' &&
        (err as ServiceError).metadata instanceof Metadata &&
        err instanceof Error;
}

function serialize(message: Message): Buffer {
    const bytes = message.serializeBinary();
    return Buffer.from(bytes.buffer, bytes.byteOffset, bytes.byteLength); // Create a Buffer view to avoid copying
}

function deserializeEvaluateResponse(bytes: Uint8Array): EvaluateResponse {
    return EvaluateResponse.deserializeBinary(bytes);
}

function deserializeEndorseResponse(bytes: Uint8Array): EndorseResponse {
    return EndorseResponse.deserializeBinary(bytes);
}

function deserializeSubmitResponse(bytes: Uint8Array): SubmitResponse {
    return SubmitResponse.deserializeBinary(bytes);
}

function deserializeCommitStatusResponse(bytes: Uint8Array): CommitStatusResponse {
    return CommitStatusResponse.deserializeBinary(bytes);
}

function deserializeChaincodeEventsResponse(bytes: Uint8Array): ChaincodeEventsResponse {
    return ChaincodeEventsResponse.deserializeBinary(bytes);
}

function deserializeDeliverResponse(bytes: Uint8Array): DeliverResponse {
    return DeliverResponse.deserializeBinary(bytes);
}

export function newGatewayClient(client: GatewayGrpcClient, defaultOptions: DefaultCallOptions): GatewayClient {
    return new GatewayClientImpl(client, defaultOptions);
}
