package org.hyperledger.fabric.client;

public class InMemoryCheckpointerTest extends CommonCheckpointerTest {

    @Override
    protected InMemoryCheckpointer getCheckpointerInstance() {
        return new InMemoryCheckpointer();
    }

    @Override
    protected void tearDown() {
    }
}
