/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions, ClientUnaryCall, requestCallback } from '@grpc/grpc-js';
import { Message } from 'google-protobuf';
import { newGatewayError } from './gatewayerror';
import { ChaincodeEventsResponse, CommitStatusResponse, EndorseRequest, EndorseResponse, EvaluateRequest, EvaluateResponse, SignedChaincodeEventsRequest, SignedCommitStatusRequest, SubmitRequest, SubmitResponse } from './protos/gateway/gateway_pb';

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

function defaultCallOptions(): CallOptions {
    return {};
}

class GatewayClientImpl implements GatewayClient {
    #client: GatewayGrpcClient;
    #defaultOptions: () => CallOptions;

    constructor(client: GatewayGrpcClient, defaultOptions: () => CallOptions = defaultCallOptions) {
        this.#client = client;
        this.#defaultOptions = defaultOptions;
    }

    evaluate(request: EvaluateRequest, options?: Readonly<CallOptions>): Promise<EvaluateResponse> {
        return new Promise((resolve, reject) => this.#client.makeUnaryRequest(
            evaluateMethod,
            serialize,
            deserializeEvaluateResponse,
            request,
            this.#buildOptions(options),
            newUnaryCallback(resolve, reject)
        ));
    }

    endorse(request: EndorseRequest, options?: Readonly<CallOptions>): Promise<EndorseResponse> {
        return new Promise((resolve, reject) => this.#client.makeUnaryRequest(
            endorseMethod,
            serialize,
            deserializeEndorseResponse,
            request,
            this.#buildOptions(options),
            newUnaryCallback(resolve, reject)
        ));
    }

    submit(request: SubmitRequest, options?: Readonly<CallOptions>): Promise<SubmitResponse> {
        return new Promise((resolve, reject) => this.#client.makeUnaryRequest(
            submitMethod,
            serialize,
            deserializeSubmitResponse,
            request,
            this.#buildOptions(options),
            newUnaryCallback(resolve, reject)
        ));
    }

    commitStatus(request: SignedCommitStatusRequest, options?: Readonly<CallOptions>): Promise<CommitStatusResponse> {
        return new Promise((resolve, reject) => this.#client.makeUnaryRequest(
            commitStatusMethod,
            serialize,
            deserializeCommitStatusResponse,
            request,
            this.#buildOptions(options),
            newUnaryCallback(resolve, reject)
        ));
    }

    chaincodeEvents(request: SignedChaincodeEventsRequest, options?: Readonly<CallOptions>): CloseableAsyncIterable<ChaincodeEventsResponse> {
        const serverStream = this.#client.makeServerStreamRequest(
            chaincodeEventsMethod,
            serialize,
            deserializeChaincodeEventsResponse,
            request,
            this.#buildOptions(options)
        );
        return {
            [Symbol.asyncIterator]: () => serverStream[Symbol.asyncIterator](),
            close: () => serverStream.cancel(),
        }
    }

    #buildOptions(options?: Readonly<CallOptions>): CallOptions {
        return Object.assign({}, this.#defaultOptions(), options);
    }
}

function newUnaryCallback<T>(resolve: (value: T) => void, reject: (reason: Error) => void): requestCallback<T> {
    return (err, value) => {
        if (err) {
            return reject(newGatewayError(err));
        }
        if (value == null) {
            return reject(new Error('No result returned'));
        }
        return resolve(value);
    }
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

export function newGatewayClient(client: GatewayGrpcClient): GatewayClient {
    return new GatewayClientImpl(client);
}
