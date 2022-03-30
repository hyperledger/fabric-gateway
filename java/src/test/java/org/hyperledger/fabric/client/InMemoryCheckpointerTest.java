package org.hyperledger.fabric.client;


import java.io.IOException;

public class InMemoryCheckpointerTest extends CommonCheckpointerTest {
    InMemoryCheckpointer inMemoryCheckpointer;

    @Override
    protected InMemoryCheckpointer getCheckpointerInstance() throws IOException {
        this.inMemoryCheckpointer = new InMemoryCheckpointer();
        return inMemoryCheckpointer;
    }

    @Override
    protected void tearDown() throws IOException {
    }

}
