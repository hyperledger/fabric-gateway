package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.peer.ChaincodeEventPackage;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

public abstract  class CommonCheckpointerTest {

    Checkpointer checkpointerInstance;

    void assertData(Checkpointer checkpointer, long blockNumber, Optional<String> transactionId) {
        assertThat(checkpointer.getBlockNumber()).isEqualTo(blockNumber);
        if (transactionId.isPresent()) {
            assertTrue(checkpointer.getTransactionId().isPresent());
            assertThat(checkpointer.getTransactionId().get()).isEqualTo(transactionId.get());
        } else {
            assertFalse(checkpointer.getTransactionId().isPresent());
        }
    }

    @BeforeEach
    void beforeEach() throws IOException {
        checkpointerInstance = getCheckpointerInstance();
    }

    @AfterEach
    void afterEach() throws IOException {
        tearDown();
    }

    protected abstract Checkpointer getCheckpointerInstance() throws IOException;
    protected abstract void tearDown() throws IOException;

    @Test
    void initial_checkpointer_state() throws IOException {
        assertData(checkpointerInstance , 0, Optional.empty());
    }

    @Test
    void CheckpointBlock_sets_next_block_number_and_empty_transaction_id() throws IOException {
        long blockNumber = 101;

        checkpointerInstance.checkpointBlock(blockNumber);
        assertData(checkpointerInstance ,blockNumber +1,Optional.empty());
    }

    @Test
    void checkpointTransaction_sets_block_number_and_transaction_id() throws Exception {
        long blockNumber = 101;

        checkpointerInstance.checkpointTransaction(blockNumber,Optional.ofNullable("txn1"));
        assertData(checkpointerInstance , blockNumber , Optional.ofNullable("txn1"));
    }

    @Test
    void checkpointEvent_sets_block_number_and_transaction_id_from_event() throws Exception {
        long blockNumber = 0;
        ChaincodeEventPackage.ChaincodeEvent event = ChaincodeEventPackage.ChaincodeEvent.newBuilder()
                .setChaincodeId("CHAINCODE_NAME")
                .setTxId("tx1")
                .setEventName("event1")
                .setPayload(ByteString.copyFromUtf8("payload1"))
                .build();
        ChaincodeEvent eventImp = new ChaincodeEventImpl(blockNumber,event);
        checkpointerInstance.checkpointChaincodeEvent(eventImp);
        assertData(checkpointerInstance ,blockNumber, Optional.ofNullable(eventImp.getTransactionId()));
    }

}
