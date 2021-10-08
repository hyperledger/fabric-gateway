/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Client, requestCallback, ServiceError } from '@grpc/grpc-js';
import { Message } from 'google-protobuf';
import { ErrorDetail, GatewayError } from './gateway';
import { ChaincodeEventsResponse, CommitStatusResponse, EndorseRequest, EndorseResponse, ErrorDetail as ErrorDetailProto, EvaluateRequest, EvaluateResponse, SignedChaincodeEventsRequest, SignedCommitStatusRequest, SubmitRequest, SubmitResponse } from './protos/gateway/gateway_pb';
import { Status } from './protos/google/rpc/status_pb';

const servicePath = '/gateway.Gateway/';
const evaluateMethod = servicePath + 'Evaluate';
const endorseMethod = servicePath + 'Endorse';
const submitMethod = servicePath + 'Submit';
const commitStatusMethod = servicePath + 'CommitStatus';
const chaincodeEventsMethod = servicePath + 'ChaincodeEvents';

export interface GatewayClient {
    evaluate(request: EvaluateRequest): Promise<EvaluateResponse>;
    endorse(request: EndorseRequest): Promise<EndorseResponse>;
    submit(request: SubmitRequest): Promise<SubmitResponse>;
    commitStatus(request: SignedCommitStatusRequest): Promise<CommitStatusResponse>;
    chaincodeEvents(request: SignedChaincodeEventsRequest): CloseableAsyncIterable<ChaincodeEventsResponse>;
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

class GatewayClientImpl implements GatewayClient {
    #client: Client;

    constructor(client: Client) {
        this.#client = client;
    }

    evaluate(request: EvaluateRequest): Promise<EvaluateResponse> {
        return new Promise((resolve, reject) =>
            this.#client.makeUnaryRequest(evaluateMethod, serialize, deserializeEvaluateResponse, request, newUnaryCallback(resolve, reject))
        );
    }

    endorse(request: EndorseRequest): Promise<EndorseResponse> {
        return new Promise((resolve, reject) =>
            this.#client.makeUnaryRequest(endorseMethod, serialize, deserializeEndorseResponse, request, newUnaryCallback(resolve, reject))
        );
    }

    submit(request: SubmitRequest): Promise<SubmitResponse> {
        return new Promise((resolve, reject) =>
            this.#client.makeUnaryRequest(submitMethod, serialize, deserializeSubmitResponse, request, newUnaryCallback(resolve, reject))
        );
    }

    commitStatus(request: SignedCommitStatusRequest): Promise<CommitStatusResponse> {
        return new Promise((resolve, reject) =>
            this.#client.makeUnaryRequest(commitStatusMethod, serialize, deserializeCommitStatusResponse, request, newUnaryCallback(resolve, reject))
        );
    }

    chaincodeEvents(request: SignedChaincodeEventsRequest): CloseableAsyncIterable<ChaincodeEventsResponse> {
        const serverStream = this.#client.makeServerStreamRequest(chaincodeEventsMethod, serialize, deserializeChaincodeEventsResponse, request);
        return {
            [Symbol.asyncIterator]: () => serverStream[Symbol.asyncIterator](),
            close: () => serverStream.cancel(),
        }
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

function newGatewayError(err: ServiceError): GatewayError {
    const result: GatewayError = new Error(err.message);
    result.code = err.code;
    result.details = err.metadata?.get('grpc-status-details-bin')
        .flatMap(metadataValue => deserializeStatus(Buffer.from(metadataValue)).getDetailsList())
        .map(statusDetail => {
            const endpointError = deserializeErrorDetail(statusDetail.getValue_asU8());
            const detail: ErrorDetail = {
                address: endpointError.getAddress(),
                message: endpointError.getMessage(),
                mspId: endpointError.getMspId(),
            };
            return detail;
        });
    return result;
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

function deserializeStatus(bytes: Uint8Array): Status {
    return Status.deserializeBinary(bytes);
}

function deserializeErrorDetail(bytes: Uint8Array): ErrorDetailProto {
    return ErrorDetailProto.deserializeBinary(bytes);
}

export function newGatewayClient(client: Client): GatewayClient {
    return new GatewayClientImpl(client);
}
