/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import io.grpc.Status;
import org.hyperledger.fabric.client.CloseableIterator;
import org.hyperledger.fabric.client.GatewayRuntimeException;

import java.util.Objects;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.SynchronousQueue;
import java.util.concurrent.TimeUnit;

public final class BasicEventListener<T> implements EventListener<T> {
    private final BlockingQueue<T> eventQueue = new SynchronousQueue<>();
    private final ExecutorService executor = Executors.newSingleThreadExecutor();
    private final CloseableIterator<T> iterator;

    public BasicEventListener(final CloseableIterator<T> iterator) {
        this.iterator = iterator;

        // Start reading events immediately as Java gRPC implementation may not invoke the gRPC service until the first
        // read attempt occurs.
        executor.execute(this::readEvents);
    }

    private void readEvents() {
        try {
            iterator.forEachRemaining(event -> {
                try {
                    eventQueue.put(event);
                } catch (InterruptedException e) {
                    iterator.close();
                    Thread.currentThread().interrupt();
                }
            });
        } catch (GatewayRuntimeException e) {
            if (e.getStatus().getCode() != Status.Code.CANCELLED) {
                throw e;
            }
        }
    }

    public T next() throws InterruptedException {
        T event = eventQueue.poll(30, TimeUnit.SECONDS);
        Objects.requireNonNull(event, "timeout waiting for event");
        return event;
    }

    public void close() {
        executor.shutdownNow();
        iterator.close();
    }
}
