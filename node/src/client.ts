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
import { ChaincodeEventsResponse, CommitStatusRequest, CommitStatusResponse, EndorseRequest, EndorseResponse, EvaluateRequest, EvaluateResponse, SignedChaincodeEventsRequest, SignedCommitStatusRequest, SubmitRequest, SubmitResponse } from './protos/gateway/gateway_pb';
import { SubmitError } from './submiterror';

const servicePath = '/gateway.Gateway/';
export const evaluateMethod = servicePath + 'Evaluate';
export const endorseMethod = servicePath + 'Endorse';
export const submitMethod = servicePath + 'Submit';
export const commitStatusMethod = servicePath + 'CommitStatus';
export const chaincodeEventsMethod = servicePath + 'ChaincodeEvents';

export interface GatewayClient {
    evaluate(request: EvaluateRequest, options?: CallOptions): Promise<EvaluateResponse>;
    endorse(request: EndorseRequest, options?: CallOptions): Promise<EndorseResponse>;
    submit(request: SubmitRequest, options?: CallOptions): Promise<SubmitResponse>;
    commitStatus(request: SignedCommitStatusRequest, options?: CallOptions): Promise<CommitStatusResponse>;
    chaincodeEvents(request: SignedChaincodeEventsRequest, options?: CallOptions): CloseableAsyncIterable<ChaincodeEventsResponse>;
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
 * Subset of the grpc-js Client used by GatewayClient to aid mocking.
 */
export interface GatewayGrpcClient {
    makeUnaryRequest<RequestType, ResponseType>(method: string, serialize: (value: RequestType) => Buffer, deserialize: (value: Buffer) => ResponseType, argument: RequestType, options: CallOptions, callback: requestCallback<ResponseType>): ClientUnaryCall;
    makeServerStreamRequest<RequestType, ResponseType>(method: string, serialize: (value: RequestType) => Buffer, deserialize: (value: Buffer) => ResponseType, argument: RequestType, options?: CallOptions): ServerStreamResponse<ResponseType>;
}

type DefaultCallOptions = Pick<ConnectOptions, 'commitStatusOptions' | 'endorseOptions' | 'evaluateOptions' | 'submitOptions' | 'chaincodeEventsOptions'>;

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
            serialize,
            deserializeChaincodeEventsResponse,
            request,
            buildOptions(this.#defaultOptions.chaincodeEventsOptions, options)
        );
    }

    #makeServerStreamRequest<RequestType, ResponseType>(
        method: string,
        serialize: (value: RequestType) => Buffer,
        deserialize: (value: Buffer) => ResponseType,
        argument: RequestType,
        options?: CallOptions
    ): CloseableAsyncIterable<ResponseType> {
        try {
            const serverStream = this.#client.makeServerStreamRequest(method, serialize, deserialize, argument, options);
            return {
                [Symbol.asyncIterator]: () => wrapAsyncIterator(serverStream[Symbol.asyncIterator]()),
                close: () => serverStream.cancel(),
            }
        } catch (err) {
            if (isServiceError(err)) {
                throw newGatewayError(err);
            }
            throw err;
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
    }
}

function wrapAsyncIterator<T>(iterator: AsyncIterator<T>): AsyncIterator<T> {
    return {
        next: async (...args) => {
            try {
                return await iterator.next(...args);
            } catch (err) {
                if (isServiceError(err)) {
                    throw newGatewayError(err);
                }
                throw err;
            }
        }
    };
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

export function newGatewayClient(client: GatewayGrpcClient, defaultOptions: DefaultCallOptions): GatewayClient {
    return new GatewayClientImpl(client, defaultOptions);
}
