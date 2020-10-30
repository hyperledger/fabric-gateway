/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Signer } from './signer'
import * as grpc from '@grpc/grpc-js';
import { Network } from './network';
import { ServiceClient } from '@grpc/grpc-js/build/src/make-client';


export interface GatewayOptions {
    url: string;
    signer: Signer;
}

export class Gateway {
    readonly _signer: Signer;
    private readonly url;
    _client!: ServiceClient;

    private constructor(options: GatewayOptions) {
        this._signer = options.signer;
        this.url = options.url;
    }

    static async connect(options: GatewayOptions): Promise<Gateway> {
        const gateway = new Gateway(options);
        await gateway._connect();
        return gateway;
    }

    private async _connect() {
        const Client = grpc.makeGenericClientConstructor({}, "protos.Gateway", {})
        this._client = new Client(
          this.url,
          grpc.credentials.createInsecure()
        )
         
        // might query available channels
    }

    getNetwork(networkName: string): Network {
        return new Network(networkName, this);
    }
}
