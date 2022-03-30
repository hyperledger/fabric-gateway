package org.hyperledger.fabric.client;



import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.nio.file.Paths;

import static org.junit.jupiter.api.Assertions.assertThrows;

public class FileCheckpointerTest extends CommonCheckpointerTest {
    FileCheckpointer fileCheckpointer;
    private static final TestUtils testUtils = TestUtils.getInstance();

    @Override
    protected FileCheckpointer getCheckpointerInstance() throws IOException {
        String file = testUtils.createTempFile(".json").toString();
        this.fileCheckpointer = new FileCheckpointer(file);
        return fileCheckpointer;
    }

    @Override
    protected void tearDown() throws IOException {
        this.fileCheckpointer.close();
    }

    @Test
    void state_is_persisted() throws IOException {
        String file = testUtils.createTempFile(".json").toString();
        this.fileCheckpointer = new FileCheckpointer(file);
        this.fileCheckpointer.checkpointTransaction(Long.valueOf(1),"TRANSACTION_ID");
        this.fileCheckpointer.close();
        this.fileCheckpointer = new FileCheckpointer(file);
        this.assertData(fileCheckpointer, Long.valueOf(1),"TRANSACTION_ID");
    }

    @Test
    void block_number_zero_is_persisted_correctly() throws IOException {
         this.getCheckpointerInstance();
         this.fileCheckpointer.checkpointBlock(Long.valueOf(0));
         this.assertData(this.fileCheckpointer, Long.valueOf(1),"");
    }

    @Test
    void throws_on_unwritable_file_location() {
        String badPath = Paths.get("").resolve("/missing_directory/checkpoint.json").toString();
        assertThrows(IOException.class, () -> {
             new FileCheckpointer(badPath);
        });
    }

}