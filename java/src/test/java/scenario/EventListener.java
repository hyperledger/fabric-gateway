/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.util.Objects;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.SynchronousQueue;
import java.util.concurrent.TimeUnit;

import org.hyperledger.fabric.client.CloseableIterator;

public final class EventListener<T> implements Events<T> {
    private final BlockingQueue<T> eventQueue = new SynchronousQueue<>();
    private final Runnable close;

    public  EventListener(CloseableIterator<T> iterator) {
        close = iterator::close;

        // Start reading events immediately as Java gRPC implementation may not invoke the gRPC service until the first
        // read attempt occurs.
        CompletableFuture.runAsync(() -> iterator.forEachRemaining(event -> {
            try {
                eventQueue.put(event);
            } catch (InterruptedException e) {
                iterator.close();
            }
        }));
    }

    public T next() {
        T event = eventQueue.poll(30, TimeUnit.SECONDS);
        Objects.requireNonNull(event, "timeout waiting for event");
        return event;
    }

    public void close() {
        close.run();
    }
}

interface Events<T> {
    T next();
    void close();
  }
