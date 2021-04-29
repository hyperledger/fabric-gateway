/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { gateway } from './protos/protos';

const servicePath = '/gateway.Gateway/';
const evaluateMethod = servicePath + 'Evaluate';
const endorseMethod = servicePath + 'Endorse';
const submitMethod = servicePath + 'Submit';
const commitStatusMethod = servicePath + 'CommitStatus';

export interface GatewayClient {
    evaluate(request: gateway.IEvaluateRequest): Promise<gateway.IEvaluateResponse>;
    endorse(request: gateway.IEndorseRequest): Promise<gateway.IEndorseResponse>;
    submit(request: gateway.ISubmitRequest): Promise<gateway.ISubmitResponse>;
    commitStatus(request: gateway.ISignedCommitStatusRequest): Promise<gateway.ICommitStatusResponse>;
}

class GatewayClientImpl implements GatewayClient {
    #client: grpc.Client;

    constructor(client: grpc.Client) {
        this.#client = client;
    }

    async evaluate(request: gateway.IEvaluateRequest): Promise<gateway.IEvaluateResponse> {
        return new Promise((resolve, reject) => {
            this.#client.makeUnaryRequest(evaluateMethod, serializeEvaluateRequest, deserializeEvaluateResponse, request, (err, value) => {
                if (err) {
                    return reject(err);
                }
                if (!value) {
                    return reject('No result returned');
                }
                return resolve(value);
            })
        });
    }

    async endorse(request: gateway.IEndorseRequest): Promise<gateway.IEndorseResponse> {
        return new Promise((resolve, reject) => {
            this.#client.makeUnaryRequest(endorseMethod, serializeEndorseRequest, deserializeEndorseResponse, request, (err, value) => {
                if (err) {
                    return reject(err);
                }
                if (!value) {
                    return reject('No result returned');
                }
                return resolve(value);
            })
        });
    }

    async submit(request: gateway.ISubmitRequest): Promise<gateway.ISubmitResponse> {
        return new Promise((resolve, reject) => {
            this.#client.makeUnaryRequest(submitMethod, serializeSubmitRequest, deserializeSubmitResponse, request, async (err, value) => {
                if (err) {
                    return reject(err);
                }
                if (!value) {
                    return reject('No result returned');
                }
                return resolve(value);
            })
        });
    }

    async commitStatus(request: gateway.ISignedCommitStatusRequest): Promise<gateway.ICommitStatusResponse> {
        return new Promise((resolve, reject) => {
            this.#client.makeUnaryRequest(commitStatusMethod, serializeSignedCommitStatusRequest, deserializeCommitStatusResponse, request, async (err, value) => {
                if (err) {
                    return reject(err);
                }
                if (!value) {
                    return reject('No result returned');
                }
                return resolve(value);
            })
        });
    }
}

function serializeEvaluateRequest(message: gateway.IEvaluateRequest): Buffer {
    const bytes = gateway.EvaluateRequest.encode(message).finish();
    return Buffer.from(bytes.buffer, bytes.byteOffset, bytes.byteLength); // Create a Buffer view to avoid copying
}

function serializeEndorseRequest(message: gateway.IEndorseRequest): Buffer {
    const bytes = gateway.EndorseRequest.encode(message).finish();
    return Buffer.from(bytes.buffer, bytes.byteOffset, bytes.byteLength); // Create a Buffer view to avoid copying
}

function serializeSubmitRequest(message: gateway.ISubmitRequest): Buffer {
    const bytes = gateway.SubmitRequest.encode(message).finish();
    return Buffer.from(bytes.buffer, bytes.byteOffset, bytes.byteLength); // Create a Buffer view to avoid copying
}

function serializeSignedCommitStatusRequest(message: gateway.ISignedCommitStatusRequest): Buffer {
    const bytes = gateway.SignedCommitStatusRequest.encode(message).finish();
    return Buffer.from(bytes.buffer, bytes.byteOffset, bytes.byteLength); // Create a Buffer view to avoid copying
}

function deserializeEvaluateResponse(bytes: Uint8Array): gateway.EvaluateResponse {
    return gateway.EvaluateResponse.decode(bytes);
}

function deserializeEndorseResponse(bytes: Uint8Array): gateway.EndorseResponse {
    return gateway.EndorseResponse.decode(bytes);
}

function deserializeSubmitResponse(bytes: Uint8Array): gateway.SubmitResponse {
    return gateway.SubmitResponse.decode(bytes);
}

function deserializeCommitStatusResponse(bytes: Uint8Array): gateway.CommitStatusResponse {
    return gateway.CommitStatusResponse.decode(bytes);
}

export function newGatewayClient(client: grpc.Client): GatewayClient {
    return new GatewayClientImpl(client);
}
