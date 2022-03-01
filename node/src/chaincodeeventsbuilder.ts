/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEventsRequest, ChaincodeEventsRequestImpl } from './chaincodeeventsrequest';
import { GatewayClient } from './client';
import { ChaincodeEventsRequest as ChaincodeEventsRequestProto } from './protos/gateway/gateway_pb';
import { SeekNextCommit, SeekPosition, SeekSpecified } from './protos/orderer/ab_pb';
import { SigningIdentity } from './signingidentity';
import { CheckPointer } from './checkpointer'
/**
 * Options used when requesting chaincode events.
 */
export interface ChaincodeEventsOptions {
    /**
     * Block number at which to start reading chaincode events.
     */
    startBlock?: bigint;
    checkPointer?: CheckPointer
}

export interface ChaincodeEventsBuilderOptions extends ChaincodeEventsOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    channelName: string;
    chaincodeName: string;
}

export class ChaincodeEventsBuilder {
    readonly #options: Readonly<ChaincodeEventsBuilderOptions>;

    constructor(options: Readonly<ChaincodeEventsBuilderOptions>) {
        this.#options = options;
    }

    async build(): Promise<ChaincodeEventsRequest> {
        return new ChaincodeEventsRequestImpl({
            client: this.#options.client,
            signingIdentity: this.#options.signingIdentity,
            request: await this.#newChaincodeEventsRequestProto(),
        });
    }

   async  #newChaincodeEventsRequestProto(): Promise<ChaincodeEventsRequestProto> {
        const result = new ChaincodeEventsRequestProto();
        result.setChannelId(this.#options.channelName);
        result.setChaincodeId(this.#options.chaincodeName);
        result.setIdentity(this.#options.signingIdentity.getCreator());
        result.setStartPosition(await this.#getStartPosition());

        return result;
    }

    async #getStartPosition(): Promise<SeekPosition> {
        const result = new SeekPosition();
        if(this.#options !== undefined ){

            const specified = new SeekSpecified();
            if(this.#options.checkPointer !== undefined) {

                const currentBlock = await this.#options.checkPointer?.getBlockNumber();
                    if(currentBlock) {

                            specified.setNumber(Number(currentBlock));
                            result.setSpecified(specified);
                            return result;
                    }


            }
            if(this.#options.startBlock){
                specified.setNumber(Number(this.#options.startBlock));
                result.setSpecified(specified);
                return result;
            }
    }
    result.setNextCommit(new SeekNextCommit());
    return result;
}
}