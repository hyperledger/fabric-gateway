/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { common, orderer } from '@hyperledger/fabric-protos';
import { Timestamp } from 'google-protobuf/google/protobuf/timestamp_pb';
import {
    BlockAndPrivateDataEventsRequest,
    BlockAndPrivateDataEventsRequestImpl,
    BlockEventsRequest,
    BlockEventsRequestImpl,
    FilteredBlockEventsRequest,
    FilteredBlockEventsRequestImpl,
} from './blockeventsrequest';
import { GatewayClient } from './client';
import { EventsBuilder, EventsOptions } from './eventsbuilder';
import { SigningIdentity } from './signingidentity';

function seekLargestBlockNumber(): orderer.SeekPosition {
    const largestBlockNumber = new orderer.SeekSpecified();
    largestBlockNumber.setNumber(Number.MAX_SAFE_INTEGER);

    const result = new orderer.SeekPosition();
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
    tlsCertificateHash?: Uint8Array;
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

export class BlockAndPrivateDataEventsBuilder {
    readonly #options: Readonly<BlockEventsBuilderOptions>;
    readonly #envelopeBuilder: BlockEventsEnvelopeBuilder;

    constructor(options: Readonly<BlockEventsBuilderOptions>) {
        this.#options = options;
        this.#envelopeBuilder = new BlockEventsEnvelopeBuilder(options);
    }

    build(): BlockAndPrivateDataEventsRequest {
        const request = this.#envelopeBuilder.build();
        return new BlockAndPrivateDataEventsRequestImpl({
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

    build(): common.Envelope {
        const result = new common.Envelope();
        result.setPayload(this.#newPayload().serializeBinary());
        return result;
    }

    #newPayload(): common.Payload {
        const result = new common.Payload();
        result.setHeader(this.#newHeader());
        result.setData(this.#newSeekInfo().serializeBinary());
        return result;
    }

    #newHeader(): common.Header {
        const result = new common.Header();
        result.setChannelHeader(this.#newChannelHeader().serializeBinary());
        result.setSignatureHeader(this.#newSignatureHeader().serializeBinary());
        return result;
    }

    #newChannelHeader(): common.ChannelHeader {
        const result = new common.ChannelHeader();
        result.setChannelId(this.#options.channelName);
        result.setEpoch(0);
        result.setTimestamp(Timestamp.fromDate(new Date()));
        result.setType(common.HeaderType.DELIVER_SEEK_INFO);
        if (this.#options.tlsCertificateHash) {
            result.setTlsCertHash(this.#options.tlsCertificateHash);
        }
        return result;
    }

    #newSignatureHeader(): common.SignatureHeader {
        const result = new common.SignatureHeader();
        result.setCreator(this.#options.signingIdentity.getCreator());
        return result;
    }

    #newSeekInfo(): orderer.SeekInfo {
        const result = new orderer.SeekInfo();
        result.setStart(this.#eventsBuilder.getStartPosition());
        result.setStop(seekLargestBlockNumber());
        return result;
    }
}
