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

class ResponseObserver<RequestType, ResponseType> implements ClientResponseObserver<RequestType, ResponseType> {
    private final AtomicReference<ClientCallStreamObserver<RequestType>> requestObserver = new AtomicReference<>();
    private final TransferQueue<ValueOrThrowable<ResponseType>> responseQueue = new LinkedTransferQueue<>();

    @Override
    public void beforeStart(final ClientCallStreamObserver<RequestType> clientCallStreamObserver) {
        requestObserver.set(clientCallStreamObserver);
    }

    @Override
    public void onNext(final ResponseType response) {
        try {
            responseQueue.transfer(ValueOrThrowable.of(response));
        } catch (InterruptedException e) {
            onError(e);
            Thread.currentThread().interrupt(); // Flag the interrupt to be handled by gRPC
        }
    }

    @Override
    public void onError(final Throwable t) {
        responseQueue.offer(ValueOrThrowable.of(t));
    }

    @Override
    public void onCompleted() {
        responseQueue.offer(ValueOrThrowable.of(new NoSuchElementException()));
    }

    public CloseableIterator<ResponseType> iterator() {
        return new CloseableIterator<ResponseType>() {
            private ValueOrThrowable<ResponseType> next;
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

            private synchronized ValueOrThrowable<ResponseType> peekNext() {
                if (null == next) {
                    try {
                        next = responseQueue.take();
                    } catch (InterruptedException e) {
                        next = ValueOrThrowable.of(e);
                    }
                }

                return next;
            }

            private synchronized ValueOrThrowable<ResponseType> readNext() {
                ValueOrThrowable<ResponseType> result = peekNext();
                next = null;
                return result;
            }

            @Override
            public ResponseType next() {
                if (closed) {
                    throw new NoSuchElementException();
                }

                ValueOrThrowable<ResponseType> response = readNext();
                ResponseType value = response.getValue();

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
