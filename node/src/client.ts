/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { protos } from './protos/protos';
import { RPCImpl } from 'protobufjs';

const serviceName = 'protos.Gateway/';

export interface GatewayClient {
    endorse(request: protos.IProposedTransaction): Promise<protos.IPreparedTransaction>;
    submit(request: protos.IPreparedTransaction): Promise<protos.IEvent>;
    evaluate(request: protos.IProposedTransaction): Promise<protos.IResult>;
}

export function newGatewayClient(client: grpc.Client): GatewayClient {
    const rpcImpl: RPCImpl = (method, requestData, callback) =>
        client.makeUnaryRequest(serviceName + method.name, bytes => bytes as Buffer, bytes => bytes, requestData, callback);
    return protos.Gateway.create(rpcImpl);
}
