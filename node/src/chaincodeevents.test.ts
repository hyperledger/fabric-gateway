/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as Long from 'long';
import { ChaincodeEvent } from './chaincodeevent';
import { ChaincodeEventCallback } from './chaincodeeventsrequest';
import { MockGatewayClient, newMockGatewayClient } from './client.test';
import { Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';
import { gateway, gateway as gatewayProto, protos } from './protos/protos';

function assertDecodeChaincodeEventsRequest(signedRequest: gatewayProto.ISignedChaincodeEventsRequest): gatewayProto.IChaincodeEventsRequest {
    expect(signedRequest.request).toBeDefined();
    return gatewayProto.ChaincodeEventsRequest.decode(signedRequest.request!); // eslint-disable-line @typescript-eslint/no-non-null-assertion
}

function newChaincodeEvent(blockNumber: number, event: protos.IChaincodeEvent): ChaincodeEvent {
    return {
        blockNumber: Long.fromInt(blockNumber),
        chaincodeId: event.chaincode_id ?? '',
        eventName: event.event_name ?? '',
        transactionId: event.tx_id ?? '',
        payload: event.payload ?? new Uint8Array(),
    };
}

function newAsyncIterable<T>(values: T[]): AsyncIterable<T> {
    return {
        async* [Symbol.asyncIterator]() { // eslint-disable-line @typescript-eslint/require-await
            for (const value of values) {
                yield value;
            }
        }
    }
}

function mockAsyncListener<T>(expectedCallCount = 1): {
    listener: (event: T) => Promise<void>,
    mock: jest.Mock<Promise<void>, T[]>,
    complete: Promise<void>,
} {
    let resolve: () => void;
    const complete = new Promise<void>(_resolve => resolve = _resolve);

    const mock = jest.fn<Promise<void>, T[]>();
    const listener = async (event: T): Promise<void> => {
        try {
            await mock(event);
        } finally {
            expectedCallCount--;
            if (expectedCallCount === 0) {
                resolve();
            }
        }
    };
    return { listener, mock, complete }
}

describe('Chaincode Events', () => {
    const channelName = 'CHANNEL_NAME';
    const signature = Buffer.from('SIGNATURE');
    const noOpListener: ChaincodeEventCallback = async () => {
        // Ignore
    };
    let client: MockGatewayClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;

    beforeEach(() => {
        client = newMockGatewayClient();
        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        }
        signer = jest.fn(undefined);
        signer.mockResolvedValue(signature);
        hash = jest.fn(undefined);
        hash.mockReturnValue(Buffer.from('DIGEST'));

        const options: InternalConnectOptions = {
            identity,
            signer,
            hash,
            gatewayClient: client,
        };
        gateway = internalConnect(options);
        network = gateway.getNetwork(channelName);
    });

    it('throws on connection error', async () => {
        client.chaincodeEvents.mockImplementation(() => {
            throw new Error('CONNECTION_ERROR');
        });

        await expect(network.onChaincodeEvent('CHAINCODE', noOpListener))
            .rejects
            .toThrow('CONNECTION_ERROR');
    });

    it('sends valid request', async () => {
        await network.onChaincodeEvent('CHAINCODE', noOpListener);

        const signedRequest = client.chaincodeEvents.mock.calls[0][0];
        expect(signedRequest.signature).toEqual(signature);

        const request = assertDecodeChaincodeEventsRequest(signedRequest);
        const expectedRequest: gatewayProto.IChaincodeEventsRequest = {
            channel_id: channelName,
            chaincode_id: 'CHAINCODE',
        };
        expect(request).toMatchObject(expectedRequest);
    });

    it('returns events', async () => {
        const event1: protos.IChaincodeEvent = {
            chaincode_id: 'CHAINCODE',
            tx_id: 'tx1',
            event_name: 'event1',
            payload: new Uint8Array(Buffer.from('payload1')),
        };
        const event2: protos.IChaincodeEvent = {
            chaincode_id: 'CHAINCODE',
            tx_id: 'tx2',
            event_name: 'event2',
            payload: new Uint8Array(Buffer.from('payload2')),
        };
        const event3: protos.IChaincodeEvent = {
            chaincode_id: 'CHAINCODE',
            tx_id: 'tx3',
            event_name: 'event3',
            payload: new Uint8Array(Buffer.from('payload3')),
        };

        const responses: gateway.IChaincodeEventsResponse[] = [
            {
                block_number: 1,
                events: [ event1, event2],
            },
            {
                block_number: 2,
                events: [ event3 ],
            },
        ];
        client.chaincodeEvents.mockReturnValue(newAsyncIterable(responses));

        const expectedEvents: ChaincodeEvent[] = [
            newChaincodeEvent(1, event1),
            newChaincodeEvent(1, event2),
            newChaincodeEvent(2, event3),
        ];

        const { listener, mock, complete } = mockAsyncListener<ChaincodeEvent>(expectedEvents.length);
        void network.onChaincodeEvent('CHAINCODE', listener);

        await complete;
        expect(mock.mock.calls.map(call => call[0])).toEqual(expectedEvents);
    });

    it('listener callback error does not stop event delivery', async () => {
        const event: protos.IChaincodeEvent = {
            chaincode_id: 'CHAINCODE',
            tx_id: 'tx',
            event_name: 'event',
            payload: new Uint8Array(Buffer.from('payload')),
        };
        const responses: gateway.IChaincodeEventsResponse[] = [
            {
                block_number: 1,
                events: [ event, event ],
            },
        ];
        client.chaincodeEvents.mockReturnValue(newAsyncIterable(responses));

        const { listener, mock, complete } = mockAsyncListener<ChaincodeEvent>(2);
        mock.mockRejectedValueOnce(new Error('EXPECTED_ERROR'));

        void network.onChaincodeEvent('CHAINCODE', listener);

        await complete;
        expect(mock).toBeCalledTimes(2);
    });
});
