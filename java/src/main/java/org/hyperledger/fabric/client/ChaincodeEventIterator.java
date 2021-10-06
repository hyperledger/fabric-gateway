/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.NoSuchElementException;

import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;

final class ChaincodeEventIterator implements CloseableIterator<ChaincodeEvent> {
    private final CloseableIterator<ChaincodeEventsResponse> responseIter;
    private ChaincodeEventsResponse currentResponse;
    private int eventIndex;
    private boolean closed = false;

    ChaincodeEventIterator(final CloseableIterator<ChaincodeEventsResponse> responseIter) {
        this.responseIter = responseIter;
    }

    @Override
    public boolean hasNext() {
        return hasNextEvent() || responseIter.hasNext();
    }

    @Override
    public ChaincodeEvent next() {
        if (closed) {
            throw new NoSuchElementException();
        }

        ChaincodeEventsResponse response = nextResponse();
        return new ChaincodeEventImpl(response.getBlockNumber(), response.getEvents(eventIndex++));
    }

    private boolean hasNextEvent() {
        return !closed && currentResponse != null && eventIndex < currentResponse.getEventsCount();
    }

    private ChaincodeEventsResponse nextResponse() {
        if (!hasNextEvent()) {
            currentResponse = responseIter.next();
            eventIndex = 0;
        }

        return currentResponse;
    }

    @Override
    public void close() {
        closed = true;
        responseIter.close();
    }
}
