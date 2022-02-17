/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Timestamp } from 'google-protobuf/google/protobuf/timestamp_pb';
import { BlockEventsRequest, BlockEventsRequestImpl, BlockEventsWithPrivateDataRequest, BlockEventsWithPrivateDataRequestImpl, FilteredBlockEventsRequest, FilteredBlockEventsRequestImpl } from './blockeventsrequest';
import { GatewayClient } from './client';
import { EventsBuilder, EventsOptions } from './eventsbuilder';
import { ChannelHeader, Envelope, Header, HeaderType, Payload, SignatureHeader } from './protos/common/common_pb';
import { SeekInfo, SeekPosition, SeekSpecified } from './protos/orderer/ab_pb';
import { SigningIdentity } from './signingidentity';

function seekLargestBlockNumber(): SeekPosition {
    const largestBlockNumber = new SeekSpecified();
    largestBlockNumber.setNumber(Number.MAX_SAFE_INTEGER);

    const result = new SeekPosition();
    result.setSpecified(largestBlockNumber);

    return result;
}

/**
 * Options used when requesting block events.
 */
export type BlockEventsOptions = EventsOptions;

export interface BlockEventsBuilderOptions extends EventsOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    channelName: string;
}

export class BlockEventsBuilder {
    readonly #options: Readonly<BlockEventsBuilderOptions>;
    readonly #envelopeBuilder: BlockEventsEnvelopeBuilder;

    constructor(options: Readonly<BlockEventsBuilderOptions>) {
        this.#options = options;
        this.#envelopeBuilder = new BlockEventsEnvelopeBuilder(options);
    }

    build(): BlockEventsRequest {
        const request = this.#envelopeBuilder.build();
        return new BlockEventsRequestImpl({
            client: this.#options.client,
            request,
            signingIdentity: this.#options.signingIdentity,
        });
    }
}

export class FilteredBlockEventsBuilder {
    readonly #options: Readonly<BlockEventsBuilderOptions>;
    readonly #envelopeBuilder: BlockEventsEnvelopeBuilder;

    constructor(options: Readonly<BlockEventsBuilderOptions>) {
        this.#options = options;
        this.#envelopeBuilder = new BlockEventsEnvelopeBuilder(options);
    }

    build(): FilteredBlockEventsRequest {
        const request = this.#envelopeBuilder.build();
        return new FilteredBlockEventsRequestImpl({
            client: this.#options.client,
            request,
            signingIdentity: this.#options.signingIdentity,
        });
    }
}

export class BlockEventsWithPrivateDataBuilder {
    readonly #options: Readonly<BlockEventsBuilderOptions>;
    readonly #envelopeBuilder: BlockEventsEnvelopeBuilder;

    constructor(options: Readonly<BlockEventsBuilderOptions>) {
        this.#options = options;
        this.#envelopeBuilder = new BlockEventsEnvelopeBuilder(options);
    }

    build(): BlockEventsWithPrivateDataRequest {
        const request = this.#envelopeBuilder.build();
        return new BlockEventsWithPrivateDataRequestImpl({
            client: this.#options.client,
            request,
            signingIdentity: this.#options.signingIdentity,
        });
    }
}

type BlockEventsEnvelopeBuilderOptions = Omit<BlockEventsBuilderOptions, 'client'>;

class BlockEventsEnvelopeBuilder {
    readonly #options: Readonly<BlockEventsEnvelopeBuilderOptions>;
    readonly #eventsBuilder: EventsBuilder;

    constructor(options: Readonly<BlockEventsEnvelopeBuilderOptions>) {
        this.#options = options;
        this.#eventsBuilder = new EventsBuilder(options);
    }

    build(): Envelope {
        const result = new Envelope();
        result.setPayload(this.#newPayload().serializeBinary());
        return result;
    }

    #newPayload(): Payload {
        const result = new Payload();
        result.setHeader(this.#newHeader());
        result.setData(this.#newSeekInfo().serializeBinary());
        return result;
    }

    #newHeader(): Header {
        const result = new Header();
        result.setChannelHeader(this.#newChannelHeader().serializeBinary());
        result.setSignatureHeader(this.#newSignatureHeader().serializeBinary());
        return result;
    }

    #newChannelHeader(): ChannelHeader {
        const result = new ChannelHeader();
        result.setChannelId(this.#options.channelName);
        result.setEpoch(0);
        result.setTimestamp(Timestamp.fromDate(new Date()));
        result.setType(HeaderType.DELIVER_SEEK_INFO);
        return result;
    }

    #newSignatureHeader(): SignatureHeader {
        const result = new SignatureHeader();
        result.setCreator(this.#options.signingIdentity.getCreator());
        return result;
    }

    #newSeekInfo(): SeekInfo {
        const result = new SeekInfo();
        result.setStart(this.#eventsBuilder.getStartPosition());
        result.setStop(seekLargestBlockNumber());
        return result;
    }
}
