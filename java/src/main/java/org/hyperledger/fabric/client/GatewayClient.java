/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;
import java.util.List;
import java.util.function.Function;

import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.Channel;
import io.grpc.Context;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.AbstractStub;
import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;
import org.hyperledger.fabric.protos.gateway.CommitStatusRequest;
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

final class GatewayClient {
    private final GatewayGrpc.GatewayBlockingStub blockingStub;
    private final CallOptions defaultOptions;

    GatewayClient(final Channel channel, final CallOptions defaultOptions) {
        GatewayUtils.requireNonNullArgument(channel, "No connection details supplied");
        GatewayUtils.requireNonNullArgument(defaultOptions, "defaultOptions");

        this.blockingStub = GatewayGrpc.newBlockingStub(channel);
        this.defaultOptions = defaultOptions;
    }

    public EvaluateResponse evaluate(final EvaluateRequest request, final CallOption... options) throws GatewayException {
        GatewayGrpc.GatewayBlockingStub stub = applyOptions(blockingStub, defaultOptions.getEvaluate(options));
        try {
            return stub.evaluate(request);
        } catch (StatusRuntimeException e) {
            throw new GatewayException(e);
        }
    }

    public EndorseResponse endorse(final EndorseRequest request, final CallOption... options) throws EndorseException {
        GatewayGrpc.GatewayBlockingStub stub = applyOptions(blockingStub, defaultOptions.getEndorse(options));
        try {
            return stub.endorse(request);
        } catch (StatusRuntimeException e) {
            throw new EndorseException(request.getTransactionId(), e);
        }
    }

    public SubmitResponse submit(final SubmitRequest request, final CallOption... options) throws SubmitException {
        GatewayGrpc.GatewayBlockingStub stub = applyOptions(blockingStub, defaultOptions.getSubmit(options));
        try {
            return stub.submit(request);
        } catch (StatusRuntimeException e) {
            throw new SubmitException(request.getTransactionId(), e);
        }
    }

    public CommitStatusResponse commitStatus(final SignedCommitStatusRequest request, final CallOption... options) throws CommitStatusException {
        GatewayGrpc.GatewayBlockingStub stub = applyOptions(blockingStub, defaultOptions.getCommitStatus(options));
        try {
            return stub.commitStatus(request);
        } catch (StatusRuntimeException e) {
            try {
                CommitStatusRequest req = CommitStatusRequest.parseFrom(request.getRequest());
                throw new CommitStatusException(req.getTransactionId(), e);
            } catch (InvalidProtocolBufferException protoEx) {
                // Should never happen
                CommitStatusException commitEx = new CommitStatusException("", e);
                commitEx.addSuppressed(protoEx);
                throw commitEx;
            }
        }
    }

    public CloseableIterator<ChaincodeEventsResponse> chaincodeEvents(final SignedChaincodeEventsRequest request, final CallOption... options) {
        GatewayGrpc.GatewayBlockingStub stub = applyOptions(blockingStub, defaultOptions.getChaincodeEvents(options));
        return invokeServerStreamingCall(stub::chaincodeEvents, request);
    }

    private static <T extends AbstractStub<T>> T applyOptions(final T stub, final List<CallOption> options) {
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
        } catch (StatusRuntimeException e) {
            context.cancel(e);
            throw new GatewayRuntimeException(e);
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
            try {
                return iterator.hasNext();
            } catch (StatusRuntimeException e) {
                throw new GatewayRuntimeException(e);
            }
        }

        @Override
        public T next() {
            try {
                return iterator.next();
            } catch (StatusRuntimeException e) {
                throw new GatewayRuntimeException(e);
            }
        }
    }
}
