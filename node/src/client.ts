/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { protos } from './protos/protos';

const servicePath = '/protos.Gateway/';
const evaluateMethod = servicePath + 'Evaluate';
const endorseMethod = servicePath + 'Endorse';
const submitMethod = servicePath + 'Submit';

export interface GatewayClient {
    evaluate(request: protos.IProposedTransaction): Promise<protos.IResult>;
    endorse(request: protos.IProposedTransaction): Promise<protos.IPreparedTransaction>;
    submit(request: protos.IPreparedTransaction): Promise<protos.IEvent>;
}

class GatewayClientImpl implements GatewayClient {
    #client: grpc.Client;

    constructor(client: grpc.Client) {
        this.#client = client;
    }

    async evaluate(request: protos.IProposedTransaction): Promise<protos.IResult> {
        return new Promise((resolve, reject) => {
            this.#client.makeUnaryRequest(evaluateMethod, serializeProposedTransaction, deserializeResult, request, (err, value) => {
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

    async endorse(request: protos.IProposedTransaction): Promise<protos.IPreparedTransaction> {
        return new Promise((resolve, reject) => {
            this.#client.makeUnaryRequest(endorseMethod, serializeProposedTransaction, deserializePreparedTransaction, request, (err, value) => {
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

    async submit(request: protos.IPreparedTransaction): Promise<protos.IEvent> {
        const stream = this.#client.makeServerStreamRequest(submitMethod, serializePreparedTransaction, deserializeEvent, request);
        return new Promise((resolve, reject) => {
            // TODO: Fix this logic for async commit wait flow
            let result: protos.IEvent;
            stream.on('data', (data) => result = data); // Received by orderer
            stream.on('end', async () => {
                await new Promise(resolve => setTimeout(resolve, 2000)); // TODO: remove this sleep once commit notification is done
                return resolve(result);
            }); // Commit received (or error?)
            stream.on('error', reject);
        });
    }
}

function serializeProposedTransaction(message: protos.IProposedTransaction): Buffer {
    const bytes = protos.ProposedTransaction.encode(message).finish();
    return Buffer.from(bytes.buffer, bytes.byteOffset, bytes.byteLength); // Create a Buffer view to avoid copying
}

function serializePreparedTransaction(message: protos.IPreparedTransaction): Buffer {
    const bytes = protos.PreparedTransaction.encode(message).finish();
    return Buffer.from(bytes.buffer, bytes.byteOffset, bytes.byteLength); // Create a Buffer view to avoid copying
}

function deserializeResult(bytes: Uint8Array): protos.Result {
    return protos.Result.decode(bytes);
}

function deserializePreparedTransaction(bytes: Uint8Array): protos.PreparedTransaction {
    return protos.PreparedTransaction.decode(bytes);
}

function deserializeEvent(bytes: Uint8Array): protos.Event {
    return protos.Event.decode(bytes);
}

export function newGatewayClient(client: grpc.Client): GatewayClient {
    return new GatewayClientImpl(client);
}
