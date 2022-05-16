package scenario;

import java.io.IOException;

import org.hyperledger.fabric.client.CloseableIterator;

public final class CheckpointEventListener<T> implements EventListener<T> {
    private final EventListener<T> eventListener;
    private final CheckpointCall<T> checkpoint;

    @FunctionalInterface
    public interface CheckpointCall<T> {
        void accept(T event) throws IOException;
    }

    CheckpointEventListener(final CloseableIterator<T> iterator, final CheckpointCall<T> checkpoint) {
        eventListener = new BasicEventListener<>(iterator);
        this.checkpoint = checkpoint;
    }

    @Override
    public T next() throws InterruptedException, IOException {
        T event = eventListener.next();
        checkpoint.accept(event);
        return event;
    }

    @Override
    public void close() {
        eventListener.close();
    }
}
