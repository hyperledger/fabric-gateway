/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.NoSuchElementException;
import java.util.function.UnaryOperator;

import io.grpc.CallOptions;
import org.hyperledger.fabric.protos.common.Envelope;
import org.hyperledger.fabric.protos.peer.BlockAndPrivateData;
import org.hyperledger.fabric.protos.peer.DeliverResponse;

final class BlockAndPrivateDataEventsRequestImpl extends SignableBlockEventsRequest implements BlockAndPrivateDataEventsRequest {
    private final GatewayClient client;

    BlockAndPrivateDataEventsRequestImpl(final GatewayClient client, final SigningIdentity signingIdentity, final Envelope request) {
        super(signingIdentity, request);
        this.client = client;
    }

    @Override
    public CloseableIterator<BlockAndPrivateData> getEvents(final UnaryOperator<CallOptions> options) {
        Envelope request = getSignedRequest();
        CloseableIterator<DeliverResponse> responseIter = client.blockAndPrivateDataEvents(request, options);

        return new MappingCloseableIterator<>(responseIter, response -> {
            DeliverResponse.TypeCase responseType = response.getTypeCase();
            if (responseType == DeliverResponse.TypeCase.STATUS) {
                throw new NoSuchElementException("Unexpected status response: " + response.getStatus());
            }

            return response.getBlockAndPrivateData();
        });
    }
}
