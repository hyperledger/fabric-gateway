/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.Channel;
import io.grpc.Context;
import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;
import org.hyperledger.fabric.protos.gateway.CommitStatusResponse;
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.EndorseResponse;
import org.hyperledger.fabric.protos.gateway.EvaluateRequest;
import org.hyperledger.fabric.protos.gateway.EvaluateResponse;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.hyperledger.fabric.protos.gateway.SubmitResponse;

import java.util.Iterator;
import java.util.function.Function;

final class GatewayClient {
    private final GatewayGrpc.GatewayBlockingStub blockingStub;

    GatewayClient(final Channel channel) {
        this.blockingStub = GatewayGrpc.newBlockingStub(channel);
    }

    public EvaluateResponse evaluate(final EvaluateRequest request) {
        return blockingStub.evaluate(request);
    }

    public EndorseResponse endorse(final EndorseRequest request) {
        return blockingStub.endorse(request);
    }

    public SubmitResponse submit(final SubmitRequest request) {
        return blockingStub.submit(request);
    }

    public CommitStatusResponse commitStatus(final SignedCommitStatusRequest request) {
        return blockingStub.commitStatus(request);
    }

    public CloseableIterator<ChaincodeEventsResponse> chaincodeEvents(final SignedChaincodeEventsRequest request) {
        return invokeServerStreamingCall(blockingStub::chaincodeEvents, request);
    }

    private static <Request, Response> ResponseIterator<Response> invokeServerStreamingCall(
            final Function<Request, Iterator<Response>> call,
            final Request request
    ) {
        Context.CancellableContext context = Context.current().withCancellation();
        try {
            Iterator<Response> iterator = context.wrap(() -> call.apply(request)).call();
            return new ResponseIterator<>(context, iterator);
        } catch (RuntimeException e) {
            context.cancel(e);
            throw e;
        } catch (Exception e) {
            // Should never happen calling a Function
            context.cancel(e);
            throw new RuntimeException(e);
        }
    }

    private static final class ResponseIterator<T> implements CloseableIterator<T> {
        private final Context.CancellableContext context;
        private final Iterator<T> iterator;

        ResponseIterator(final Context.CancellableContext context, final Iterator<T> iterator) {
            this.context = context;
            this.iterator = iterator;
        }

        @Override
        public void close() {
            context.close();
        }

        @Override
        public boolean hasNext() {
            return iterator.hasNext();
        }

        @Override
        public T next() {
            return iterator.next();
        }
    }
}
