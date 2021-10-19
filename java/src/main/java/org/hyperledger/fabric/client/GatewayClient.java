/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;
import java.util.function.Function;

import io.grpc.Channel;
import io.grpc.Context;
import io.grpc.stub.AbstractStub;
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

// Non-final only to allow spying with Mockito
class GatewayClient {
    private final GatewayGrpc.GatewayBlockingStub blockingStub;

    GatewayClient(final Channel channel) {
        this.blockingStub = GatewayGrpc.newBlockingStub(channel);
    }

    public EvaluateResponse evaluate(final EvaluateRequest request, final CallOption... options) {
        return applyOptions(blockingStub, options).evaluate(request);
    }

    public EndorseResponse endorse(final EndorseRequest request, final CallOption... options) {
        return applyOptions(blockingStub, options).endorse(request);
    }

    public SubmitResponse submit(final SubmitRequest request, final CallOption... options) {
        return applyOptions(blockingStub, options).submit(request);
    }

    public CommitStatusResponse commitStatus(final SignedCommitStatusRequest request, final CallOption... options) {
        return applyOptions(blockingStub, options).commitStatus(request);
    }

    public CloseableIterator<ChaincodeEventsResponse> chaincodeEvents(final SignedChaincodeEventsRequest request, final CallOption... options) {
        return invokeServerStreamingCall(applyOptions(blockingStub, options)::chaincodeEvents, request);
    }

    private static <T extends AbstractStub<T>> T applyOptions(final T stub, final CallOption... options) {
        T result = stub;
        for (CallOption option : options) {
            result = option.apply(result);
        }

        return result;
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
