/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { gateway } from '@hyperledger/fabric-protos';
import { ChaincodeEventsRequest, ChaincodeEventsRequestImpl } from './chaincodeeventsrequest';
import { GatewayClient } from './client';
import { EventsBuilder, EventsOptions } from './eventsbuilder';
import { SigningIdentity } from './signingidentity';

/**
 * Options used when requesting chaincode events.
 */
export type ChaincodeEventsOptions = EventsOptions;

export interface ChaincodeEventsBuilderOptions extends EventsOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    channelName: string;
    chaincodeName: string;
}

export class ChaincodeEventsBuilder {
    readonly #options: Readonly<ChaincodeEventsBuilderOptions>;
    readonly #eventsBuilder: EventsBuilder;

    constructor(options: Readonly<ChaincodeEventsBuilderOptions>) {
        this.#options = options;
        this.#eventsBuilder = new EventsBuilder(options);
    }

    build(): ChaincodeEventsRequest {
        return new ChaincodeEventsRequestImpl({
            client: this.#options.client,
            signingIdentity: this.#options.signingIdentity,
            signedRequest: this.#newSignedChaincodeEventsRequestProto(),
        });
    }

    #newChaincodeEventsRequestProto(): gateway.ChaincodeEventsRequest {
        const result = new gateway.ChaincodeEventsRequest();
        result.setChannelId(this.#options.channelName);
        result.setChaincodeId(this.#options.chaincodeName);
        result.setIdentity(this.#options.signingIdentity.getCreator());
        result.setStartPosition(this.#eventsBuilder.getStartPosition());
        const transactionId = this.#options.checkpoint?.getTransactionId();
        if (transactionId) {
            result.setAfterTransactionId(transactionId);
        }
        return result;
    }

    #newSignedChaincodeEventsRequestProto(): gateway.SignedChaincodeEventsRequest {
        const result = new gateway.SignedChaincodeEventsRequest();
        result.setRequest(this.#newChaincodeEventsRequestProto().serializeBinary());
        return result;
    }
}
