/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;

import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;

final class ChaincodeEventIterator implements Iterator<ChaincodeEvent> {
    private final Iterator<ChaincodeEventsResponse> responseIter;
    private ChaincodeEventsResponse currentResponse;
    private int eventIndex;

    ChaincodeEventIterator(final Iterator<ChaincodeEventsResponse> responseIter) {
        this.responseIter = responseIter;
    }

    @Override
    public boolean hasNext() {
        return hasNextEvent() || responseIter.hasNext();
    }

    @Override
    public ChaincodeEvent next() {
        ChaincodeEventsResponse response = nextResponse();
        return new ChaincodeEvent(response.getBlockNumber(), response.getEvents(eventIndex++));
    }

    private boolean hasNextEvent() {
        return currentResponse != null && eventIndex < currentResponse.getEventsCount();
    }

    private ChaincodeEventsResponse nextResponse() {
        if (!hasNextEvent()) {
            currentResponse = responseIter.next();
            eventIndex = 0;
        }

        return currentResponse;
    }
}
