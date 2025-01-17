/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { connect, ConnectOptions, Identity } from '../src';

const utf8Encoder = new TextEncoder();

describe('Gateway', () => {
    let identity: Identity;
    let connectOptions: ConnectOptions;

    beforeEach(() => {
        identity = {
            mspId: 'MSP_ID',
            credentials: utf8Encoder.encode('CERTIFICATE'),
        };
        connectOptions = {
            identity,
            signer: (message) => Promise.resolve(message),
        };
    });

    describe('connect', () => {
        it('throws if no identity supplied', () => {
            const options = Object.assign(connectOptions, { identity: undefined });
            expect(() => connect(options)).toThrow();
        });
        it('throws if no signer supplied', () => {
            const options = Object.assign(connectOptions, { signer: undefined });
            expect(() => connect(options)).toThrow();
        });
    });

    describe('getNetwork', () => {
        it('returns correctly named network', () => {
            const gateway = connect(connectOptions);

            const network = gateway.getNetwork('CHANNEL_NAME');

            expect(network.getName()).toBe('CHANNEL_NAME');
        });
    });

    describe('getIdentity', () => {
        it('returns supplied identity', () => {
            const gateway = connect(connectOptions);

            const result = gateway.getIdentity();

            expect(result.mspId).toEqual(identity.mspId);
            expect(new Uint8Array(result.credentials)).toEqual(new Uint8Array(identity.credentials));
        });
    });
});
