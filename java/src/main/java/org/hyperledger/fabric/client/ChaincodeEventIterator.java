/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Collections;
import java.util.Iterator;

import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;

final class ChaincodeEventIterator implements CloseableIterator<ChaincodeEvent> {
    private final CloseableIterator<ChaincodeEventsResponse> responseIter;
    private Iterator<org.hyperledger.fabric.protos.peer.ChaincodeEvent> eventIter = Collections.emptyIterator();
    private long blockNumber;

    ChaincodeEventIterator(final CloseableIterator<ChaincodeEventsResponse> responseIter) {
        this.responseIter = responseIter;
    }

    @Override
    public boolean hasNext() {
        return eventIter.hasNext() || responseIter.hasNext();
    }

    @Override
    public ChaincodeEvent next() {
        while (!eventIter.hasNext()) {
            ChaincodeEventsResponse response = responseIter.next();
            eventIter = response.getEventsList().iterator();
            blockNumber = response.getBlockNumber();
        }

        return new ChaincodeEventImpl(blockNumber, eventIter.next());
    }

    @Override
    public void close() {
        responseIter.close();
    }
}
