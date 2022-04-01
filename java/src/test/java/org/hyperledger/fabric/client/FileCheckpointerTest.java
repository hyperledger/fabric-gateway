package org.hyperledger.fabric.client;



import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.assertThrows;

public class FileCheckpointerTest extends CommonCheckpointerTest {
    FileCheckpointer fileCheckpointer;
    private static final TestUtils testUtils = TestUtils.getInstance();

    @Override
    protected FileCheckpointer getCheckpointerInstance() throws IOException {
        Path file = testUtils.createTempFile(".json");
        this.fileCheckpointer = new FileCheckpointer(file);
        return fileCheckpointer;
    }

    @Override
    protected void tearDown() throws IOException {
        this.fileCheckpointer.close();
    }

    @Test
    void state_is_persisted() throws IOException {
        Path file = testUtils.createTempFile(".json");
        this.fileCheckpointer = new FileCheckpointer(file);
        this.fileCheckpointer.checkpointTransaction(1, Optional.ofNullable("TRANSACTION_ID"));
        this.fileCheckpointer.close();
        this.fileCheckpointer = new FileCheckpointer(file);
        this.assertData(fileCheckpointer, 1, Optional.ofNullable("TRANSACTION_ID"));
    }

    @Test
    void block_number_zero_is_persisted_correctly() throws IOException {
         this.getCheckpointerInstance();
         this.fileCheckpointer.checkpointBlock(0);
         this.assertData(this.fileCheckpointer, 1, Optional.empty());
    }

    @Test
    void throws_on_unwritable_file_location() {
        Path badPath = Paths.get("").resolve("/missing_directory/checkpoint.json");
        assertThrows(IOException.class, () -> {
             new FileCheckpointer(badPath);
        });
    }

}