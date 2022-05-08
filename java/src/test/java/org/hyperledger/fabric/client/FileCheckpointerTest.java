package org.hyperledger.fabric.client;


import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;

import static org.junit.jupiter.api.Assertions.assertThrows;

public class FileCheckpointerTest extends CommonCheckpointerTest {
    private FileCheckpointer checkpointer;
    private static final TestUtils testUtils = TestUtils.getInstance();

    @Override
    protected FileCheckpointer getCheckpointerInstance() throws IOException {
        Path file = testUtils.createTempFile(".json");
        checkpointer = new FileCheckpointer(file);
        return checkpointer;
    }

    @Override
    protected void tearDown() throws IOException {
        checkpointer.close();
    }

    @Test
    void state_is_persisted() throws IOException {
        Path file = testUtils.createTempFile(".json");
        try (FileCheckpointer checkpointer = new FileCheckpointer(file)) {
            checkpointer.checkpointTransaction(1, "TRANSACTION_ID");
        }
        checkpointer = new FileCheckpointer(file);
        assertCheckpoint(checkpointer, 1, "TRANSACTION_ID");
    }

    @Test
    void partial_state_is_persisted() throws IOException {
        Path file = testUtils.createTempFile(".json");
        try (FileCheckpointer checkpointer = new FileCheckpointer(file)) {
            checkpointer.checkpointBlock(1);
        }
        checkpointer = new FileCheckpointer(file);
        assertCheckpoint(checkpointer, 2);
    }

    @Test
    void block_number_zero_is_persisted_correctly() throws IOException {
         getCheckpointerInstance();
         checkpointer.checkpointBlock(0);
         assertCheckpoint(this.checkpointer, 1);
    }

    @Test
    void throws_on_unwritable_file_location() {
        Path badPath = Paths.get("").resolve("/missing_directory/checkpoint.json");
        assertThrows(IOException.class, () -> new FileCheckpointer(badPath));
    }
}
