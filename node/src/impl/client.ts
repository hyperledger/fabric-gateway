/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import * as grpc from '@grpc/grpc-js';
import { ServiceClient } from '@grpc/grpc-js/build/src/make-client';
import { protos } from '../protos/protos'

export interface Client {
    _evaluate(signedProposal: protos.IProposedTransaction): Promise<string>;
    _endorse(signedProposal: protos.IProposedTransaction): Promise<protos.IPreparedTransaction>;
    _submit(preparedTransaction: protos.IPreparedTransaction): Promise<protos.IEvent>
}

export class ClientImpl implements Client {
    private readonly serviceClient: ServiceClient;

    constructor(url: string) {
        const SvcClient = grpc.makeGenericClientConstructor({}, "protos.Gateway", {})
        this.serviceClient = new SvcClient(
          url,
          grpc.credentials.createInsecure()
        )
    }

    private rpcSimple(method: any, requestData: any, callback: any) {
        this.serviceClient.makeUnaryRequest(
            'protos.Gateway/' + method.name,
            arg => arg,
            arg => arg,
            requestData,
            new grpc.Metadata(),
            {},
            callback
        )
    }

    private rpcStream(method: any, requestData: any, callback: any) {
        let data: any = undefined;
        const call = this.serviceClient.makeServerStreamRequest(
            'protos.Gateway/' + method.name,
            arg => arg,
            arg => arg,
            requestData,
            new grpc.Metadata(),
            {}
        );
        call.on('data', function (event: any) {
            data = event;
        });
        call.on('end', function () {
            callback(null, data);
        });
        call.on('error', function (e: any) {
            // An error has occurred and the stream has been closed.
            callback(e, null);
        });
        call.on('status', function (status: any) {
            // process status
        });
    }

    async _evaluate(signedProposal: protos.IProposedTransaction): Promise<string> {
        const service = protos.Gateway.create(this.rpcSimple.bind(this), false, false);
        return new Promise((resolve, reject) => {
            service.evaluate(signedProposal, function (err: Error|null, result: protos.IResult|undefined) {
                if (err) reject(err);
                resolve(result?.value?.toString());
            });
        })
    }

    async _endorse(signedProposal: protos.IProposedTransaction): Promise<protos.IPreparedTransaction> {
        const service = protos.Gateway.create(this.rpcSimple.bind(this), false, false);
        // const endorse:any  = promisify(service.endorse.bind(this));
        // return endorse(signedProposal);
        return new Promise((resolve, reject) => {
            service.endorse(signedProposal, function (err: Error|null, result: protos.IPreparedTransaction|undefined) {
                if (err) reject(err);
                resolve(result);
            });
        })
    }

    async _submit(preparedTransaction: protos.IPreparedTransaction): Promise<protos.IEvent> {
        const service = protos.Gateway.create(this.rpcStream.bind(this), false, false);
        return new Promise((resolve, reject) => {
            service.submit(preparedTransaction, function (err: Error|null, result: protos.IEvent|undefined) {
                if (err) reject(err);
                resolve(result);
            });
        })
    }

}