package org.hyperledger.fabric.client;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public final class CheckpointerTest {

    @BeforeEach
    void beforeEach() {

    }

    @AfterEach
    void afterEach() {

    }

    @Test
    void sets_next_block_number_and_empty_transaction_id() {
        assertThatThrownBy(() -> network.getChaincodeEvents(null))
                .isInstanceOf(NullPointerException.class)
                .hasMessageContaining("chaincode name");
    }
    @Test
    void sets_block_number_and_transaction_id(){

    }

    @Test
    void sets_block_number_and_transaction_id_from_event(){

    }
}
