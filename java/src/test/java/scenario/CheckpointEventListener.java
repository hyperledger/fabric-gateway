package scenario;

import org.hyperledger.fabric.client.CloseableIterator;

import java.util.function.Consumer;

public final class CheckpointEventListener<T> implements Events<T> {
  private final EventListener<T> eventListener;
  private final Consumer<T> checkpoint;

  CheckpointEventListener(CloseableIterator<T> iterator, Consumer<T> checkpoint) {
    eventListener = new EventListener<T>(iterator);
    this.checkpoint = checkpoint;
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
