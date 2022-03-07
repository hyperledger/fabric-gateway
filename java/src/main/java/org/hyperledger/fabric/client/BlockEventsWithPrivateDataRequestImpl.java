/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.NoSuchElementException;

import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.peer.EventsPackage;

final class BlockEventsWithPrivateDataRequestImpl extends SignableBlockEventsRequest implements BlockEventsWithPrivateDataRequest {
    private final GatewayClient client;

    BlockEventsWithPrivateDataRequestImpl(final GatewayClient client, final SigningIdentity signingIdentity, final Common.Envelope request) {
        super(signingIdentity, request);
        this.client = client;
    }

    @Override
    public CloseableIterator<EventsPackage.BlockAndPrivateData> getEvents(final CallOption... options) {
        Common.Envelope request = getSignedRequest();
        CloseableIterator<EventsPackage.DeliverResponse> responseIter = client.blockEventsWithPrivateData(request, options);

        return new MappingCloseableIterator<>(responseIter, response -> {
            EventsPackage.DeliverResponse.TypeCase responseType = response.getTypeCase();
            if (responseType == EventsPackage.DeliverResponse.TypeCase.STATUS) {
                throw new NoSuchElementException("Unexpected status response: " + response.getStatus());
            }

            return response.getBlockAndPrivateData();
        });
    }
}
