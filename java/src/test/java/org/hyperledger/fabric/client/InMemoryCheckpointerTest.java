package org.hyperledger.fabric.client;


import java.io.IOException;

public class InMemoryCheckpointerTest extends CommonCheckpointerTest {

    @Override
    protected InMemoryCheckpointer getCheckpointerInstance() throws IOException {
        return new InMemoryCheckpointer();
    }

    @Override
    protected void tearDown() throws IOException {
    }

}
