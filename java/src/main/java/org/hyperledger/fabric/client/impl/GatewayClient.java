/*
 *  Copyright 2020 IBM All Rights Reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.util.Iterator;

import org.hyperledger.fabric.gateway.Event;
import org.hyperledger.fabric.gateway.PreparedTransaction;
import org.hyperledger.fabric.gateway.ProposedTransaction;
import org.hyperledger.fabric.gateway.Result;

public interface GatewayClient extends AutoCloseable {
    PreparedTransaction endorse(ProposedTransaction request);
    Iterator<Event> submit(PreparedTransaction request);
    Result evaluate(ProposedTransaction request);
    @Override
    void close();
}
