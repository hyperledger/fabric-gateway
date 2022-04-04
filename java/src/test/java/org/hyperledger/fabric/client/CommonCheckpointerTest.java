package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.peer.ChaincodeEventPackage;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertTrue;


public abstract  class CommonCheckpointerTest {

    private Checkpointer checkpointerInstance;

    protected void assertCheckpoint(final Checkpoint checkpoint, final long blockNumber) {
        assertThat(checkpoint.getBlockNumber()).isEqualTo(blockNumber);
        assertThat(checkpoint.getTransactionId()).isNotPresent();
    }
    protected void assertCheckpoint(final Checkpoint checkpoint, final long blockNumber, final String transactionId) {
        assertThat(checkpoint.getBlockNumber()).isEqualTo(blockNumber);
        assertTrue(checkpoint.getTransactionId().isPresent());
        assertThat(checkpoint.getTransactionId().get()).isEqualTo(transactionId);
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
        assertCheckpoint(checkpointerInstance , 0);
    }

    @Test
    void checkpointBlock_sets_next_block_number_and_empty_transaction_id() throws IOException {
        checkpointerInstance.checkpointBlock(101);
        assertCheckpoint(checkpointerInstance , 102);
    }

    @Test
    void checkpointTransaction_sets_block_number_and_transaction_id() throws Exception {
        checkpointerInstance.checkpointTransaction(101, "txn1");
        assertCheckpoint(checkpointerInstance , 101 , "txn1");
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
        assertCheckpoint(checkpointerInstance ,blockNumber, eventImp.getTransactionId());
    }

}
