package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.peer.ChaincodeEventPackage;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;

import static org.assertj.core.api.Assertions.assertThat;

public abstract  class CommonCheckpointerTest {

    Checkpointer checkpointerInstance;

    void assertData(Checkpointer checkpointer, Long blockNumber, String transactionId){
        assertThat(checkpointer.getBlockNumber()).isEqualTo(blockNumber);
        assertThat(checkpointer.getTransactionId()).isEqualTo(transactionId);
    }

    @BeforeEach
    void beforeEach() throws IOException {
        this.checkpointerInstance = getCheckpointerInstance();
    }

    @AfterEach
    void afterEach() throws IOException {
        tearDown();
    }

    protected abstract Checkpointer getCheckpointerInstance() throws IOException;
    protected abstract void tearDown() throws IOException;

    @Test
    void initial_checkpointer_state() throws IOException {
        assertData(this.checkpointerInstance ,Long.valueOf(0),"");
    }

    @Test
    void CheckpointBlock_sets_next_block_number_and_empty_transaction_id() throws IOException {
        Long blockNumber = Long.valueOf(101);

        this.checkpointerInstance.checkpointBlock(blockNumber);
        assertData(this.checkpointerInstance ,blockNumber +1,"");
    }

    @Test
    void checkpointTransaction_sets_block_number_and_transaction_id() throws Exception {
        Long blockNumber = Long.valueOf(101);

        this.checkpointerInstance.checkpointTransaction(blockNumber,"txn1");
        assertData(this.checkpointerInstance ,blockNumber ,"txn1");
    }

    @Test
    void checkpointiEvent_sets_block_number_and_transaction_id_from_event() throws Exception {
        Long blockNumber = Long.valueOf(0);
        ChaincodeEventPackage.ChaincodeEvent event = ChaincodeEventPackage.ChaincodeEvent.newBuilder()
                .setChaincodeId("CHAINCODE_NAME")
                .setTxId("tx1")
                .setEventName("event1")
                .setPayload(ByteString.copyFromUtf8("payload1"))
                .build();
        ChaincodeEvent eventImp = new ChaincodeEventImpl(blockNumber,event);
        this.checkpointerInstance.checkpointChaincodeEvent(eventImp);
        assertData(this.checkpointerInstance ,blockNumber ,eventImp.getTransactionId());
    }

}
