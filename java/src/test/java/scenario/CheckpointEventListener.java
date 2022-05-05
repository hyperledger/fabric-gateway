package scenario;

import org.hyperledger.fabric.client.CloseableIterator;

import java.util.function.Consumer;

public final class CheckpointEventListener<T> implements Events<T> {
  EventListener<T> eventListener;
  Consumer<T> checkpoint;
  CheckpointEventListener(CloseableIterator<T> iterator, Consumer<T> checkpoint) {
    eventListener = new EventListener<T>(iterator);
    checkpoint = checkpoint;
  }
  @Override
  public T next() {
    T event = eventListener.next();
    checkpoint.accept(event);
    return event;
  }
  @Override
  public void close() {
    eventListener.close();
  }
}
