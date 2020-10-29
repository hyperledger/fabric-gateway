/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Signer } from './signer'
import { protosGateway } from './impl/protoutils'
import * as grpc from '@grpc/grpc-js';
import { Network } from './network';

export interface GatewayOptions {
    url: string;
    signer: Signer;
}

export class Gateway {
    readonly _signer: Signer;
    readonly _stub: any;

    private constructor(options: GatewayOptions) {
        this._signer = options.signer;
        this._stub = new protosGateway(options.url, grpc.credentials.createInsecure());
    }

    static async connect(options: GatewayOptions): Promise<Gateway> {
        const gateway = new Gateway(options);
        await gateway._connect();
        return gateway;
    }

    private async _connect() {
        // might query available channels
    }

    getNetwork(networkName: string): Network {
        return new Network(networkName, this);
    }
}
