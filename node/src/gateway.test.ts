/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { connect, ConnectOptions } from './gateway';
import { Identity } from './identity/identity';

describe('Gateway', () => {
    let identity: Identity;
    let client: grpc.Client;

    beforeEach(() => {
        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        };
        client = new grpc.Client('example.org:1337', grpc.credentials.createInsecure());
    });

    describe('connect', () => {
        it('throws if no client connection supplied', () => {
            const options = {
                identity,
            } as ConnectOptions;
            expect(() => connect(options)).toThrow();
        });

        it('throws if no identity supplied', () => {
            const options = {
                client,
            } as unknown as ConnectOptions;
            expect(() => connect(options)).toThrow();
        });
    });

    describe('getNetwork', () => {
        it('returns correctly named network', () => {
            const options: ConnectOptions = {
                identity,
                client,
            };
            const gateway = connect(options);

            const network = gateway.getNetwork('CHANNEL_NAME');

            expect(network.getName()).toBe('CHANNEL_NAME');
        });
    });

    describe('getIdentity', () => {
        it('returns supplied identity', () => {
            const options: ConnectOptions = {
                identity,
                client,
            };
            const gateway = connect(options);

            const result = gateway.getIdentity();

            expect(result.mspId).toEqual(identity.mspId);
            expect(Uint8Array.from(result.credentials)).toEqual(Uint8Array.from(identity.credentials));
        });
    });

    describe('close', () => {
        it('does not close supplied gRPC client', () => {
            const client = new grpc.Client('example.org:1337', grpc.credentials.createInsecure());
            const closeStub = (client.close = jest.fn());
            const options: ConnectOptions = {
                identity,
                client,
            };
            const gateway = connect(options);

            gateway.close();

            expect(closeStub).not.toHaveBeenCalled();
        });

        it('called by resource clean-up', () => {
            const client = new grpc.Client('example.org:1337', grpc.credentials.createInsecure());
            const options: ConnectOptions = {
                identity,
                client,
            };
            const closeStub = jest.fn();
            {
                // @ts-expect-error Assigned to unused variable for resource cleanup
                // eslint-disable-next-line @typescript-eslint/no-unused-vars
                using gateway = Object.assign(connect(options), { close: closeStub });
            }

            expect(closeStub).toHaveBeenCalled();
        });
    });
});
