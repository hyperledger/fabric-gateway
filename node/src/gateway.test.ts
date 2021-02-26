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

    beforeEach(() => {
        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        }
    });

    describe('connect', () => {
        it('throws if no connection details supplied', async () => {
            const options: ConnectOptions = {
                identity,
            };
            await expect(connect(options)).rejects.toBeDefined();
        });

        it('throws if no identity supplied', async () => {
            const options: ConnectOptions = {
                url: 'example.org:1337',
            } as ConnectOptions;
            await expect(connect(options)).rejects.toBeDefined();
        });

        it('using endpoint address', async () => {
            const options: ConnectOptions = {
                identity,
                url: 'example.org:1337',
            } as ConnectOptions;
            await expect(connect(options)).resolves.toBeDefined();
        });

        it('using TLS endpoint address', async () => {
            const options: ConnectOptions = {
                identity,
                url: 'example.org:1337',
                tlsRootCertificates: Buffer.from("pem"),
                serverNameOverride: "localhost"
            } as ConnectOptions;
            await expect(connect(options)).resolves.toBeDefined();
        });

        it('using gRPC client', async () => {
            const Client = grpc.makeGenericClientConstructor({}, '');
            const client = new Client('example.org:1337', grpc.credentials.createInsecure());
            const options: ConnectOptions = {
                identity,
                client,
            };
            await expect(connect(options)).resolves.toBeDefined();
        });
    });

    describe('getNetwork', () => {
        it('returns correctly named network', async () => {
            const options: ConnectOptions = {
                identity,
                url: 'example.org:1337',
            };
            const gateway = await connect(options);

            const network = gateway.getNetwork('CHANNEL_NAME');

            expect(network.getName()).toBe('CHANNEL_NAME');
        });
    });

    describe('getIdentity', () => {
        it('returns supplied identity', async () => {
            const options: ConnectOptions = {
                identity,
                url: 'example.org:1337',
            };
            const gateway = await connect(options);

            const result = gateway.getIdentity();

            expect(result.mspId).toEqual(identity.mspId);
            expect(Uint8Array.from(result.credentials)).toEqual(Uint8Array.from(identity.credentials));
        });
    });

    describe('close', () => {
        it('does not close supplied gRPC client', async() => {
            const Client = grpc.makeGenericClientConstructor({}, '');
            const client = new Client('example.org:1337', grpc.credentials.createInsecure());
            client.close = jest.fn();
            const options: ConnectOptions = {
                identity,
                client,
            };
            const gateway = await connect(options);

            gateway.close();

            expect(client.close).not.toHaveBeenCalled();
        });
    });
});
