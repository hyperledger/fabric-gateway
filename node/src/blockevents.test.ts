/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions, Metadata, ServiceError, status } from '@grpc/grpc-js';
import { common, ledger, msp, orderer, peer } from '@hyperledger/fabric-protos';
import { CloseableAsyncIterable } from '.';
import { BlockEventsOptions } from './blockeventsbuilder';
import { BlockAndPrivateDataEventsRequest, BlockEventsRequest, FilteredBlockEventsRequest } from './blockeventsrequest';
import { assertDefined, Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { GatewayError } from './gatewayerror';
import { Identity } from './identity/identity';
import { Network } from './network';
import { DuplexStreamResponseStub, MockGatewayGrpcClient, newDuplexStreamResponse, readElements } from './testutils.test';
import * as checkpointers from './checkpointers';

function assertStartPositionToBeSpecified(seekInfo: orderer.SeekInfo, blockNumber: number): void {
    const start = seekInfo.getStart();
    expect(start).toBeDefined();
    expect(start?.getTypeCase()).toBe(orderer.SeekPosition.TypeCase.SPECIFIED);
    expect(start?.getSpecified()).toBeDefined();
    expect(start?.getSpecified()?.getNumber()).toBe(blockNumber);
}

function assertStartPositionToBeNextCommit(seekInfo: orderer.SeekInfo): void {
    const start = seekInfo.getStart();
    expect(start).toBeDefined();
    expect(start?.getTypeCase()).toBe(orderer.SeekPosition.TypeCase.NEXT_COMMIT);
    expect(start?.getNextCommit()).toBeDefined();
}

function assertStopPosition(seekInfo: orderer.SeekInfo): void {
    const stop = seekInfo.getStop();
    expect(stop).toBeDefined();
    expect(stop?.getTypeCase()).toBe(orderer.SeekPosition.TypeCase.SPECIFIED);
    expect(stop?.getSpecified()).toBeDefined();
    expect(stop?.getSpecified()?.getNumber()).toBe(Number.MAX_SAFE_INTEGER);
}

describe('Block Events', () => {
    const channelName = 'CHANNEL_NAME';
    const signature = Buffer.from('SIGNATURE');
    const serviceError: ServiceError = Object.assign(new Error('ERROR_MESSAGE'), {
        code: status.UNAVAILABLE,
        details: 'DETAILS',
        metadata: new Metadata(),
    });

    let defaultOptions: () => CallOptions;
    let client: MockGatewayGrpcClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;

    beforeEach(() => {
        const now = new Date();
        const callOptions = {
            deadline: now.setHours(now.getHours() + 1),
        };
        defaultOptions = () => callOptions; // Return a specific object to test modification

        client = new MockGatewayGrpcClient();
        identity = {
            mspId: 'MSP_ID',
            credentials: new Uint8Array(Buffer.from('CERTIFICATE')),
        };
        signer = jest.fn(undefined);
        signer.mockResolvedValue(signature);
        hash = jest.fn(undefined);
        hash.mockReturnValue(Buffer.from('DIGEST'));

        const options: InternalConnectOptions = {
            identity,
            signer,
            hash,
            client,
            blockEventsOptions: defaultOptions,
            filteredBlockEventsOptions: defaultOptions,
            blockAndPrivateDataEventsOptions: defaultOptions,
        };
        gateway = internalConnect(options);
        network = gateway.getNetwork(channelName);
    });

    function assertValidBlockEventsRequestHeader(payload: common.Payload): void {
        const header = assertDefined(payload.getHeader(), 'header');
        const channelHeader = common.ChannelHeader.deserializeBinary(header.getChannelHeader_asU8());
        const signatureHeader = common.SignatureHeader.deserializeBinary(header.getSignatureHeader_asU8());
        const creator = msp.SerializedIdentity.deserializeBinary(signatureHeader.getCreator_asU8());

        const actualCreator: Identity = {
            credentials: creator.getIdBytes_asU8(),
            mspId: creator.getMspid(),
        };

        expect(channelHeader.getChannelId()).toBe(network.getName());
        expect(actualCreator).toEqual(gateway.getIdentity());
    }

    interface TestCase {
        description: string;
        mockResponse(stream: DuplexStreamResponseStub<common.Envelope, peer.DeliverResponse>): void;
        mockError(err: ServiceError): void;
        getEvents(options?: BlockEventsOptions): Promise<CloseableAsyncIterable<unknown>>;
        newEventsRequest(options?: BlockEventsOptions): BlockEventsRequest | FilteredBlockEventsRequest | BlockAndPrivateDataEventsRequest;
        getCallOptions(): CallOptions[];
        newBlockResponse(blockNumber: number): peer.DeliverResponse;
        getBlockFromResponse(response: peer.DeliverResponse): common.Block | peer.FilteredBlock | peer.BlockAndPrivateData | undefined;
    }

    const testCases: TestCase[] = [
        {
            description: 'Blocks',
            mockResponse(stream) {
                client.mockBlockEventsResponse(stream);
            },
            mockError(err) {
                client.mockBlockEventsError(err);
            },
            getEvents(options?) {
                return network.getBlockEvents(options);
            },
            newEventsRequest(options?) {
                return network.newBlockEventsRequest(options);
            },
            getCallOptions() {
                return client.getBlockEventsOptions();
            },
            newBlockResponse(blockNumber) {
                const header = new common.BlockHeader();
                header.setNumber(blockNumber);

                const block = new common.Block();
                block.setHeader(header);

                const response = new peer.DeliverResponse();
                response.setBlock(block);

                return response;
            },
            getBlockFromResponse(response) {
                return response.getBlock();
            },
        },
        {
            description: 'Filtered blocks',
            mockResponse(stream) {
                client.mockFilteredBlockEventsResponse(stream);
            },
            mockError(err) {
                client.mockFilteredBlockEventsError(err);
            },
            getEvents(options?) {
                return network.getFilteredBlockEvents(options);
            },
            newEventsRequest(options?) {
                return network.newFilteredBlockEventsRequest(options);
            },
            getCallOptions() {
                return client.getFilteredBlockEventsOptions();
            },
            newBlockResponse(blockNumber) {
                const block = new peer.FilteredBlock();
                block.setNumber(blockNumber);

                const response = new peer.DeliverResponse();
                response.setFilteredBlock(block);

                return response;
            },
            getBlockFromResponse(response) {
                return response.getFilteredBlock();
            },
        },
        {
            description: 'Blocks and private data',
            mockResponse(stream) {
                client.mockBlockAndPrivateDataEventsResponse(stream);
            },
            mockError(err) {
                client.mockBlockAndPrivateDataEventsError(err);
            },
            getEvents(options?) {
                return network.getBlockAndPrivateDataEvents(options);
            },
            newEventsRequest(options?) {
                return network.newBlockAndPrivateDataEventsRequest(options);
            },
            getCallOptions() {
                return client.getBlockAndPrivateDataEventsOptions();
            },
            newBlockResponse(blockNumber) {
                const header = new common.BlockHeader();
                header.setNumber(blockNumber);

                const block = new common.Block();
                block.setHeader(header);

                const blockAndPrivateData = new peer.BlockAndPrivateData();
                blockAndPrivateData.setBlock(block);
                const privateData = blockAndPrivateData.getPrivateDataMapMap();
                privateData.set(0, new ledger.rwset.TxPvtReadWriteSet());

                const response = new peer.DeliverResponse();
                response.setBlockAndPrivateData(blockAndPrivateData);

                return response;
            },
            getBlockFromResponse(response) {
                return response.getBlockAndPrivateData();
            },
        },
    ];
    testCases.forEach(testCase => describe(`${testCase.description}`, () => {
        it('sends valid request with default start position', async () => {
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>([]);
            testCase.mockResponse(stream);

            await testCase.getEvents();

            expect(stream.write.mock.calls.length).toBe(1);
            const request = stream.write.mock.calls[0][0];

            const payload = common.Payload.deserializeBinary(request.getPayload_asU8());
            assertValidBlockEventsRequestHeader(payload);

            const seekInfo = orderer.SeekInfo.deserializeBinary(payload.getData_asU8());

            assertStartPositionToBeNextCommit(seekInfo);
            assertStopPosition(seekInfo);
        });

        it('throws with negative specified start block number', async () => {
            const startBlock = BigInt(-1);
            await expect(network.getBlockEvents({ startBlock }))
                .rejects
                .toThrow();
        });

        it('sends valid request with specified start block number', async () => {
            const startBlock = BigInt(418);
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>([]);
            testCase.mockResponse(stream);

            await testCase.getEvents({ startBlock });

            expect(stream.write.mock.calls.length).toBe(1);
            const request = stream.write.mock.calls[0][0];

            const payload = common.Payload.deserializeBinary(request.getPayload_asU8());
            assertValidBlockEventsRequestHeader(payload);

            const seekInfo = orderer.SeekInfo.deserializeBinary(payload.getData_asU8());

            assertStartPositionToBeSpecified(seekInfo, Number(startBlock));
            assertStopPosition(seekInfo);
        });

        it('Uses specified start block instead of unset checkpoint', async () => {
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>([]);
            testCase.mockResponse(stream);
            const startBlock = BigInt(418);
            await testCase.getEvents({startBlock: startBlock, checkpoint: checkpointers.inMemory()});

            expect(stream.write.mock.calls.length).toBe(1);
            const request = stream.write.mock.calls[0][0];

            const payload = common.Payload.deserializeBinary(request.getPayload_asU8());
            assertValidBlockEventsRequestHeader(payload);

            const seekInfo = orderer.SeekInfo.deserializeBinary(payload.getData_asU8());

            assertStartPositionToBeSpecified(seekInfo, Number(startBlock));
            assertStopPosition(seekInfo);
        });

        it('Uses checkpoint block instead of specified start block', async () => {
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>([]);
            testCase.mockResponse(stream);

            const startBlock = BigInt(418);
            const checkpointer = checkpointers.inMemory();
            await checkpointer.checkpointBlock(1n);
            await testCase.getEvents({startBlock: startBlock, checkpoint: checkpointer});

            expect(stream.write.mock.calls.length).toBe(1);
            const request = stream.write.mock.calls[0][0];

            const payload = common.Payload.deserializeBinary(request.getPayload_asU8());
            assertValidBlockEventsRequestHeader(payload);

            const seekInfo = orderer.SeekInfo.deserializeBinary(payload.getData_asU8());

            assertStartPositionToBeSpecified(seekInfo, 2);
            assertStopPosition(seekInfo);
        });

        it('Uses checkpoint block zero with set transaction ID instead of specified start block', async () => {
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>([]);
            testCase.mockResponse(stream);

            const startBlock = BigInt(418);
            const blockNumber = 0n;
            const checkpointer = checkpointers.inMemory();
            await checkpointer.checkpointTransaction(blockNumber, 'transactionID');
            await testCase.getEvents({startBlock: startBlock, checkpoint: checkpointer});

            expect(stream.write.mock.calls.length).toBe(1);
            const request = stream.write.mock.calls[0][0];

            const payload = common.Payload.deserializeBinary(request.getPayload_asU8());
            assertValidBlockEventsRequestHeader(payload);

            const seekInfo = orderer.SeekInfo.deserializeBinary(payload.getData_asU8());

            assertStartPositionToBeSpecified(seekInfo, Number(blockNumber));
            assertStopPosition(seekInfo);
        });

        it('Uses default start block instead of unset checkpoint and no start block', async () => {
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>([]);
            testCase.mockResponse(stream);
            const checkpointer = checkpointers.inMemory();
            await testCase.getEvents({checkpoint: checkpointer});

            expect(stream.write.mock.calls.length).toBe(1);
            const request = stream.write.mock.calls[0][0];

            const payload = common.Payload.deserializeBinary(request.getPayload_asU8());
            assertValidBlockEventsRequestHeader(payload);

            const seekInfo = orderer.SeekInfo.deserializeBinary(payload.getData_asU8());

            assertStartPositionToBeNextCommit(seekInfo);
            assertStopPosition(seekInfo);
        });

        it('uses specified call options', async () => {
            const deadline = Date.now() + 1000;

            await testCase.newEventsRequest()
                .getEvents({ deadline });

            const actual = testCase.getCallOptions()[0];
            expect(actual.deadline).toBe(deadline);
        });

        it('uses default call options', async () => {
            await testCase.getEvents();

            const actual = testCase.getCallOptions()[0];
            expect(actual.deadline).toBe(defaultOptions().deadline);
        });

        it('default call options are not modified', async () => {
            const expected = defaultOptions().deadline;
            const deadline = Date.now() + 1000;

            await network.newBlockEventsRequest()
                .getEvents({ deadline });

            expect(defaultOptions().deadline).toBe(expected);
        });

        it('throws GatewayError on call ServiceError', async () => {
            testCase.mockError(serviceError);

            const t = testCase.getEvents();

            await expect(t).rejects.toThrow(GatewayError);
            await expect(t).rejects.toThrow(serviceError.message);
            await expect(t).rejects.toMatchObject({
                code: serviceError.code,
                cause: serviceError,
            });
        });

        it('throws GatewayError on receive ServiceError', async () => {
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>([serviceError]);
            testCase.mockResponse(stream);

            const events = await testCase.getEvents();
            const t = readElements(events, 1);

            await expect(t).rejects.toThrow(GatewayError);
            await expect(t).rejects.toThrow(serviceError.message);
            await expect(t).rejects.toMatchObject({
                code: serviceError.code,
                cause: serviceError,
            });
        });

        it('throws on receive of status message', async () => {
            const response = new peer.DeliverResponse();
            response.setStatus(common.Status.SERVICE_UNAVAILABLE);
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>([response]);
            testCase.mockResponse(stream);

            const events = await testCase.getEvents();
            const t = readElements(events, 1);

            await expect(t).rejects.toThrow(String(common.Status.SERVICE_UNAVAILABLE));
        });

        it('throws on receive of unexpected message type', async () => {
            const response = new peer.DeliverResponse();
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>([response]);
            testCase.mockResponse(stream);

            const events = await testCase.getEvents();
            const t = readElements(events, 1);

            await expect(t).rejects.toThrow(String(response.getTypeCase()));
        });

        it('receives events', async () => {
            const responses = [testCase.newBlockResponse(1), testCase.newBlockResponse(2)];
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>(responses);
            testCase.mockResponse(stream);

            const events = await testCase.getEvents();
            const actual = await readElements(events, responses.length);

            const expected = responses.map(response => testCase.getBlockFromResponse(response));

            expect(actual).toEqual(expected);
        });

        it('closing iterator cancels gRPC stream', async () => {
            const responses = [testCase.newBlockResponse(1), testCase.newBlockResponse(2)];
            const stream = newDuplexStreamResponse<common.Envelope, peer.DeliverResponse>(responses);
            testCase.mockResponse(stream);

            const events = await testCase.getEvents();
            events.close();

            expect(stream.cancel).toBeCalled();
        });
    }));
});
