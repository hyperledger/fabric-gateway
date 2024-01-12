/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions, ClientUnaryCall, Metadata, requestCallback, ServiceError } from '@grpc/grpc-js';
import { common, gateway, peer } from '@hyperledger/fabric-protos';
import { Message } from 'google-protobuf';
import { CommitStatusError } from './commitstatuserror';
import { EndorseError } from './endorseerror';
import { ConnectOptions } from './gateway';
import { GatewayError, newGatewayError } from './gatewayerror';
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
    evaluate(request: gateway.EvaluateRequest, options?: CallOptions): Promise<gateway.EvaluateResponse>;
    endorse(request: gateway.EndorseRequest, options?: CallOptions): Promise<gateway.EndorseResponse>;
    submit(request: gateway.SubmitRequest, options?: CallOptions): Promise<gateway.SubmitResponse>;
    commitStatus(
        request: gateway.SignedCommitStatusRequest,
        options?: CallOptions,
    ): Promise<gateway.CommitStatusResponse>;
    chaincodeEvents(
        request: gateway.SignedChaincodeEventsRequest,
        options?: CallOptions,
    ): CloseableAsyncIterable<gateway.ChaincodeEventsResponse>;
    blockEvents(request: common.Envelope, options?: CallOptions): CloseableAsyncIterable<peer.DeliverResponse>;
    filteredBlockEvents(request: common.Envelope, options?: CallOptions): CloseableAsyncIterable<peer.DeliverResponse>;
    blockAndPrivateDataEvents(
        request: common.Envelope,
        options?: CallOptions,
    ): CloseableAsyncIterable<peer.DeliverResponse>;
}

// @ts-expect-error Polyfill for Symbol.dispose if not present
Symbol.dispose ??= Symbol('Symbol.dispose');

/**
 * An async iterable that can be closed when the consumer does not want to read any more elements, freeing up resources
 * that may be held by the iterable.
 *
 * This type implements the Disposable interface, allowing instances to be disposed of with ECMAScript explicit
 * resource management and the `using` keyword instead of calling {@link close} directly.
 *
 * @see [ECMAScript explicit resource management](https://github.com/tc39/proposal-explicit-resource-management)
 */
export interface CloseableAsyncIterable<T> extends AsyncIterable<T> {
    /**
     * Close the iterable to free up resources when no more elements are required.
     */
    close(): void;

    [Symbol.dispose](): void;
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
    makeUnaryRequest<RequestType, ResponseType>(
        method: string,
        serialize: (value: RequestType) => Buffer,
        deserialize: (value: Buffer) => ResponseType,
        argument: RequestType,
        options: CallOptions,
        callback: requestCallback<ResponseType>,
    ): ClientUnaryCall;
    makeServerStreamRequest<RequestType, ResponseType>(
        method: string,
        serialize: (value: RequestType) => Buffer,
        deserialize: (value: Buffer) => ResponseType,
        argument: RequestType,
        options: CallOptions,
    ): ServerStreamResponse<ResponseType>;
    makeBidiStreamRequest<RequestType, ResponseType>(
        method: string,
        serialize: (value: RequestType) => Buffer,
        deserialize: (value: Buffer) => ResponseType,
        options: CallOptions,
    ): DuplexStreamResponse<RequestType, ResponseType>;
}

type DefaultCallOptions = Pick<
    ConnectOptions,
    | 'commitStatusOptions'
    | 'endorseOptions'
    | 'evaluateOptions'
    | 'submitOptions'
    | 'chaincodeEventsOptions'
    | 'blockEventsOptions'
    | 'filteredBlockEventsOptions'
    | 'blockAndPrivateDataEventsOptions'
>;

class GatewayClientImpl implements GatewayClient {
    readonly #client: GatewayGrpcClient;
    readonly #defaultOptions: Readonly<DefaultCallOptions>;

    constructor(client: GatewayGrpcClient, defaultOptions: Readonly<DefaultCallOptions>) {
        this.#client = client;
        this.#defaultOptions = Object.assign({}, defaultOptions);
    }

    evaluate(request: gateway.EvaluateRequest, options?: Readonly<CallOptions>): Promise<gateway.EvaluateResponse> {
        return new Promise((resolve, reject) =>
            this.#client.makeUnaryRequest(
                evaluateMethod,
                serialize,
                deserializeEvaluateResponse,
                request,
                buildOptions(this.#defaultOptions.evaluateOptions, options),
                newUnaryCallback(resolve, reject),
            ),
        );
    }

    endorse(request: gateway.EndorseRequest, options?: Readonly<CallOptions>): Promise<gateway.EndorseResponse> {
        return new Promise((resolve, reject) =>
            this.#client.makeUnaryRequest(
                endorseMethod,
                serialize,
                deserializeEndorseResponse,
                request,
                buildOptions(this.#defaultOptions.endorseOptions, options),
                newUnaryCallback(
                    resolve,
                    reject,
                    (err) => new EndorseError(Object.assign(err, { transactionId: request.getTransactionId() })),
                ),
            ),
        );
    }

    submit(request: gateway.SubmitRequest, options?: Readonly<CallOptions>): Promise<gateway.SubmitResponse> {
        return new Promise((resolve, reject) =>
            this.#client.makeUnaryRequest(
                submitMethod,
                serialize,
                deserializeSubmitResponse,
                request,
                buildOptions(this.#defaultOptions.submitOptions, options),
                newUnaryCallback(
                    resolve,
                    reject,
                    (err) => new SubmitError(Object.assign(err, { transactionId: request.getTransactionId() })),
                ),
            ),
        );
    }

    commitStatus(
        request: gateway.SignedCommitStatusRequest,
        options?: Readonly<CallOptions>,
    ): Promise<gateway.CommitStatusResponse> {
        return new Promise((resolve, reject) =>
            this.#client.makeUnaryRequest(
                commitStatusMethod,
                serialize,
                deserializeCommitStatusResponse,
                request,
                buildOptions(this.#defaultOptions.commitStatusOptions, options),
                newUnaryCallback(resolve, reject, (err) => {
                    const req = gateway.CommitStatusRequest.deserializeBinary(request.getRequest_asU8());
                    return new CommitStatusError(Object.assign(err, { transactionId: req.getTransactionId() }));
                }),
            ),
        );
    }

    chaincodeEvents(
        request: gateway.SignedChaincodeEventsRequest,
        options?: Readonly<CallOptions>,
    ): CloseableAsyncIterable<gateway.ChaincodeEventsResponse> {
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
        options: CallOptions,
    ): CloseableAsyncIterable<ResponseType> {
        try {
            const serverStream = this.#client.makeServerStreamRequest(
                method,
                serialize,
                deserialize,
                argument,
                options,
            );
            return {
                [Symbol.asyncIterator]: () => wrapAsyncIterator(serverStream[Symbol.asyncIterator]()),
                close: () => {
                    serverStream.cancel();
                },
                [Symbol.dispose]: () => {
                    serverStream.cancel();
                },
            };
        } catch (err) {
            rethrowGrpcError(err);
        }
    }

    blockEvents(request: common.Envelope, options?: CallOptions): CloseableAsyncIterable<peer.DeliverResponse> {
        return this.#makeBidiStreamRequest(
            deliverMethod,
            deserializeDeliverResponse,
            request,
            buildOptions(this.#defaultOptions.blockEventsOptions, options),
        );
    }

    filteredBlockEvents(request: common.Envelope, options?: CallOptions): CloseableAsyncIterable<peer.DeliverResponse> {
        return this.#makeBidiStreamRequest(
            deliverFilteredMethod,
            deserializeDeliverResponse,
            request,
            buildOptions(this.#defaultOptions.filteredBlockEventsOptions, options),
        );
    }

    blockAndPrivateDataEvents(
        request: common.Envelope,
        options?: CallOptions,
    ): CloseableAsyncIterable<peer.DeliverResponse> {
        return this.#makeBidiStreamRequest(
            deliverWithPrivateDataMethod,
            deserializeDeliverResponse,
            request,
            buildOptions(this.#defaultOptions.blockAndPrivateDataEventsOptions, options),
        );
    }

    #makeBidiStreamRequest<RequestType extends Message, ResponseType>(
        method: string,
        deserialize: (value: Buffer) => ResponseType,
        request: RequestType,
        options: CallOptions,
    ): CloseableAsyncIterable<ResponseType> {
        try {
            const duplexStream = this.#client.makeBidiStreamRequest(method, serialize, deserialize, options);
            duplexStream.write(request);
            return {
                [Symbol.asyncIterator]: () => wrapAsyncIterator(duplexStream[Symbol.asyncIterator]()),
                close: () => {
                    duplexStream.cancel();
                },
                [Symbol.dispose]: () => {
                    duplexStream.cancel();
                },
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
    wrap: (err: GatewayError) => GatewayError = (err) => err,
): requestCallback<T> {
    return (err, value) => {
        if (err) {
            reject(wrap(newGatewayError(err)));
            return;
        }
        if (value == null) {
            reject(new Error('No result returned'));
            return;
        }
        resolve(value);
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
        },
    };
}

function rethrowGrpcError(err: unknown): never {
    if (isServiceError(err)) {
        throw newGatewayError(err);
    }
    throw err;
}

function isServiceError(err: unknown): err is ServiceError {
    return (
        typeof (err as ServiceError).code === 'number' &&
        typeof (err as ServiceError).details === 'string' &&
        (err as ServiceError).metadata instanceof Metadata &&
        err instanceof Error
    );
}

function serialize(message: Message): Buffer {
    const bytes = message.serializeBinary();
    return Buffer.from(bytes.buffer, bytes.byteOffset, bytes.byteLength); // Create a Buffer view to avoid copying
}

function deserializeEvaluateResponse(bytes: Uint8Array): gateway.EvaluateResponse {
    return gateway.EvaluateResponse.deserializeBinary(bytes);
}

function deserializeEndorseResponse(bytes: Uint8Array): gateway.EndorseResponse {
    return gateway.EndorseResponse.deserializeBinary(bytes);
}

function deserializeSubmitResponse(bytes: Uint8Array): gateway.SubmitResponse {
    return gateway.SubmitResponse.deserializeBinary(bytes);
}

function deserializeCommitStatusResponse(bytes: Uint8Array): gateway.CommitStatusResponse {
    return gateway.CommitStatusResponse.deserializeBinary(bytes);
}

function deserializeChaincodeEventsResponse(bytes: Uint8Array): gateway.ChaincodeEventsResponse {
    return gateway.ChaincodeEventsResponse.deserializeBinary(bytes);
}

function deserializeDeliverResponse(bytes: Uint8Array): peer.DeliverResponse {
    return peer.DeliverResponse.deserializeBinary(bytes);
}

export function newGatewayClient(client: GatewayGrpcClient, defaultOptions: DefaultCallOptions): GatewayClient {
    return new GatewayClientImpl(client, defaultOptions);
}
