/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.NoSuchElementException;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.LinkedTransferQueue;
import java.util.concurrent.TransferQueue;
import java.util.concurrent.atomic.AtomicReference;

import io.grpc.stub.ClientCallStreamObserver;
import io.grpc.stub.ClientResponseObserver;

class ResponseObserver<ReqT, RespT> implements ClientResponseObserver<ReqT, RespT> {
    private final AtomicReference<ClientCallStreamObserver<ReqT>> requestObserver = new AtomicReference<>();
    private final TransferQueue<ValueOrThrowable<RespT>> responseQueue = new LinkedTransferQueue<>();

    @Override
    public void beforeStart(final ClientCallStreamObserver<ReqT> clientCallStreamObserver) {
        requestObserver.set(clientCallStreamObserver);
    }

    @Override
    public void onNext(final RespT response) {
        try {
            responseQueue.transfer(new ValueOrThrowable<>(response));
        } catch (InterruptedException e) {
            requestObserver.get().cancel("failed to deliver event", e);
            onError(e);
        }
    }

    @Override
    public void onError(final Throwable t) {
        responseQueue.offer(new ValueOrThrowable<>(t));
    }

    @Override
    public void onCompleted() {
        responseQueue.offer(new ValueOrThrowable<>(new NoSuchElementException()));
    }

    public CloseableIterator<RespT> iterator() {
        return new CloseableIterator<RespT>() {
            private ValueOrThrowable<RespT> next;
            private boolean closed = false;

            @Override
            public void close() {
                closed = true;
                // Call close async since it seems possible to block behind event delivery, at least in unit tests
                CompletableFuture.runAsync(() -> requestObserver.get().cancel("client close", null));
            }

            @Override
            public boolean hasNext() {
                return !closed && !(peekNext().getThrowable() instanceof NoSuchElementException);
            }

            private synchronized ValueOrThrowable<RespT> peekNext() {
                if (null == next) {
                    try {
                        next = responseQueue.take();
                    } catch (InterruptedException e) {
                        next = new ValueOrThrowable<>(e);
                    }
                }

                return next;
            }

            private synchronized ValueOrThrowable<RespT> readNext() {
                ValueOrThrowable<RespT> result = peekNext();
                next = null;
                return result;
            }

            @Override
            public RespT next() {
                if (closed) {
                    throw new NoSuchElementException();
                }

                ValueOrThrowable<RespT> response = readNext();
                RespT value = response.getValue();

                if (value != null) {
                    return value;
                }

                closed = true;
                Throwable t = response.getThrowable();

                if (t instanceof Error) {
                    throw (Error) t;
                } else if (t instanceof RuntimeException) {
                    throw (RuntimeException) t;
                } else {
                    NoSuchElementException e = new NoSuchElementException();
                    e.addSuppressed(t);
                    throw e;
                }
            }
        };
    }
}
