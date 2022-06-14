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
import org.hyperledger.fabric.protos.peer.DeliverResponse;
import org.hyperledger.fabric.protos.peer.FilteredBlock;

final class FilteredBlockEventsRequestImpl extends SignableBlockEventsRequest implements FilteredBlockEventsRequest {
    private final GatewayClient client;

    FilteredBlockEventsRequestImpl(final GatewayClient client, final SigningIdentity signingIdentity, final Envelope request) {
        super(signingIdentity, request);
        this.client = client;
    }

    @Override
    public CloseableIterator<FilteredBlock> getEvents(final UnaryOperator<CallOptions> options) {
        Envelope request = getSignedRequest();
        CloseableIterator<DeliverResponse> responseIter = client.filteredBlockEvents(request, options);

        return new MappingCloseableIterator<>(responseIter, response -> {
            DeliverResponse.TypeCase responseType = response.getTypeCase();
            if (responseType == DeliverResponse.TypeCase.STATUS) {
                throw new NoSuchElementException("Unexpected status response: " + response.getStatus());
            }

            return response.getFilteredBlock();
        });
    }
}
