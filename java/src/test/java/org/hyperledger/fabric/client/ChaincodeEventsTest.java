/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import io.grpc.CallOptions;
import io.grpc.Deadline;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.orderer.Ab;
import org.hyperledger.fabric.protos.peer.ChaincodeEventPackage;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.TimeUnit;
import java.util.function.Supplier;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.assertj.core.api.Assertions.catchThrowableOfType;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;

public final class ChaincodeEventsTest {
    private static final Deadline defaultDeadline = Deadline.after(1, TimeUnit.DAYS);

    private GatewayMocker mocker;
    private GatewayServiceStub stub;
    private Gateway gateway;
    private Network network;

    @BeforeEach
    void beforeEach() {
        mocker = new GatewayMocker();
        stub = mocker.getGatewayServiceStubSpy();

        gateway = mocker.getGatewayBuilder()
                .chaincodeEventsOptions(CallOption.deadline(defaultDeadline))
                .connect();
        network = gateway.getNetwork("NETWORK");
    }

    @AfterEach
    void afterEach() {
        gateway.close();
        mocker.close();
    }

    void assertStartPosition(final org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest actual, final long blockNumber) {
        assertStartPosition(actual, blockNumber, "");
    }

    void assertStartPosition(final org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest actual, final long blockNumber, final Optional<String> transactionId) {
        assertStartPosition(actual, blockNumber, transactionId.orElse(""));
    }

    void assertStartPosition(final org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest actual, final long blockNumber, final String transactionId) {
        assertThat(actual.getStartPosition().getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.SPECIFIED);
        assertThat(actual.getStartPosition().getSpecified().getNumber()).isEqualTo(blockNumber);
        assertThat(actual.getAfterTransactionId()).isEqualTo(transactionId);
    }

    void assertNextCommit(final org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest actual) {
        assertThat(actual.getStartPosition().getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.NEXT_COMMIT);
    }

    void assertRequestInitiated(final Supplier<CloseableIterator<?>> supplier) {
        try (CloseableIterator<?> iter = supplier.get()) {
            iter.hasNext(); // Interact with iterator to ensure async request has been made
        }
    }

    @Test
    void throws_NullPointerException_on_null_chaincode_name() {
        assertThatThrownBy(() -> network.getChaincodeEvents(null))
                .isInstanceOf(NullPointerException.class)
                .hasMessageContaining("chaincode name");
    }

    @Test
    void throws_on_connection_error() {
        StatusRuntimeException expected = new StatusRuntimeException(Status.UNAVAILABLE);
        doThrow(expected).when(stub).chaincodeEvents(any());

        GatewayRuntimeException e = catchThrowableOfType(
                () -> {
                    try (CloseableIterator<ChaincodeEvent> events = network.getChaincodeEvents("CHAINCODE_NAME")) {
                        events.forEachRemaining(event -> { });
                    }
                },
                GatewayRuntimeException.class
        );
        assertThat(e.getStatus()).isEqualTo(expected.getStatus());
        assertThat(e).hasCauseInstanceOf(StatusRuntimeException.class);
    }

    @Test
    void sends_valid_request_with_default_start_position() throws Exception {
        assertRequestInitiated(() -> network.getChaincodeEvents("CHAINCODE_NAME"));

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());
        assertNextCommit(request);
    }

    @Test
    void sends_valid_request_with_specified_start_block_number() throws Exception {
        long startBlock = 101;
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(startBlock)
                .build();

        assertRequestInitiated(eventsRequest::getEvents);

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());
        assertStartPosition(request, startBlock);
    }

    @Test
    void sends_valid_request_with_specified_start_block_number_using_sign_bit_for_unsigned_64bit_value() throws Exception {
        long startBlock = -1;
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(startBlock)
                .build();
        assertRequestInitiated(eventsRequest::getEvents);

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertStartPosition(request, startBlock);
    }

    @Test
    void uses_specified_start_block_instead_of_unset_checkpoint() throws Exception {
        long startBlock = -1;
        Checkpointer checkpointer = new InMemoryCheckpointer();
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(startBlock)
                .checkpoint(checkpointer)
                .build();

        assertRequestInitiated(eventsRequest::getEvents);

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertStartPosition(request, startBlock);
    }

    @Test
    void uses_checkpoint_block_instead_of_specified_start_block() throws Exception {
        Checkpointer checkpointer = new InMemoryCheckpointer();
        checkpointer.checkpointBlock(101);
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(-1)
                .checkpoint(checkpointer)
                .build();

        assertRequestInitiated(eventsRequest::getEvents);

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());
        assertStartPosition(request, checkpointer.getBlockNumber());
    }

    @Test
    void uses_checkpoint_block_and_transaction_instead_of_specified_start_block() throws Exception {
        Checkpointer checkpointer = new InMemoryCheckpointer();
        checkpointer.checkpointTransaction(101, "TRANSACTION_ID");
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(-1)
                .checkpoint(checkpointer)
                .build();

        assertRequestInitiated(eventsRequest::getEvents);

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertStartPosition(request, checkpointer.getBlockNumber(), checkpointer.getTransactionId());
    }

    @Test
    void start_at_next_commit_with_unset_checkpoint_and_no_start_block() throws Exception {
        Checkpointer checkpointer = new InMemoryCheckpointer();
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .checkpoint(checkpointer)
                .build();

        assertRequestInitiated(eventsRequest::getEvents);

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertNextCommit(request);
    }

    @Test
    void uses_checkpoint_block_and_transaction_with_unset_start_block() throws Exception {
        Checkpointer checkpointer = new InMemoryCheckpointer();
        checkpointer.checkpointTransaction(1, "TRANSACTION_ID");
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .checkpoint(checkpointer)
                .build();

        assertRequestInitiated(eventsRequest::getEvents);

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertStartPosition(request, checkpointer.getBlockNumber(), checkpointer.getTransactionId());
    }

    @Test
    void uses_checkpointed_chaincode_event_block_and_transaction() throws  Exception {
        long blockNumber = 1;
        ChaincodeEventPackage.ChaincodeEvent event = ChaincodeEventPackage.ChaincodeEvent.newBuilder()
                .setChaincodeId("CHAINCODE_NAME")
                .setTxId("tx1")
                .setEventName("event")
                .setPayload(ByteString.copyFromUtf8("payload1"))
                .build();
        ChaincodeEventImpl chaincodeEvent = new ChaincodeEventImpl(blockNumber, event);
        Checkpointer checkpointer = new InMemoryCheckpointer();

        checkpointer.checkpointChaincodeEvent(chaincodeEvent);
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .checkpoint(checkpointer)
                .build();

        assertRequestInitiated(eventsRequest::getEvents);

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertStartPosition(request, chaincodeEvent.getBlockNumber(), chaincodeEvent.getTransactionId());
    }

    @Test
    void returns_events() {
        ChaincodeEventPackage.ChaincodeEvent event1 = ChaincodeEventPackage.ChaincodeEvent.newBuilder()
                .setChaincodeId("CHAINCODE_NAME")
                .setTxId("tx1")
                .setEventName("event1")
                .setPayload(ByteString.copyFromUtf8("payload1"))
                .build();
        ChaincodeEventPackage.ChaincodeEvent event2 = ChaincodeEventPackage.ChaincodeEvent.newBuilder()
                .setChaincodeId("CHAINCODE_NAME")
                .setTxId("tx2")
                .setEventName("event2")
                .setPayload(ByteString.copyFromUtf8("payload2"))
                .build();
        ChaincodeEventPackage.ChaincodeEvent event3 = ChaincodeEventPackage.ChaincodeEvent.newBuilder()
                .setChaincodeId("CHAINCODE_NAME")
                .setTxId("tx3")
                .setEventName("event3")
                .setPayload(ByteString.copyFromUtf8("payload3"))
                .build();

        Stream<ChaincodeEventsResponse> responses = Stream.of(
                ChaincodeEventsResponse.newBuilder()
                        .setBlockNumber(1)
                        .addEvents(event1)
                        .addEvents(event2)
                        .build(),
                ChaincodeEventsResponse.newBuilder()
                        .setBlockNumber(2)
                        .addEvents(event3)
                        .build()
        );
        doReturn(responses).when(stub).chaincodeEvents(any());

        try (CloseableIterator<ChaincodeEvent> actual = network.getChaincodeEvents("CHAINCODE_NAME")) {
            List<ChaincodeEvent> expected = Arrays.asList(
                    new ChaincodeEventImpl(1, event1),
                    new ChaincodeEventImpl(1, event2),
                    new ChaincodeEventImpl(2, event3)
            );
            assertThat(actual)
                    .toIterable()
                    .hasSameElementsAs(expected);
        }
    }

    @Test
    void close_stops_receiving_events() {
        ChaincodeEventPackage.ChaincodeEvent event1 = ChaincodeEventPackage.ChaincodeEvent.newBuilder()
                .setChaincodeId("CHAINCODE_NAME")
                .setTxId("tx1")
                .setEventName("event1")
                .setPayload(ByteString.copyFromUtf8("payload1"))
                .build();
        ChaincodeEventsResponse response = ChaincodeEventsResponse.newBuilder()
                .setBlockNumber(1)
                .addEvents(event1)
                .build();

        Stream<ChaincodeEventsResponse> responses = Stream.generate(() -> response);
        doReturn(responses).when(stub).chaincodeEvents(any());

        CloseableIterator<ChaincodeEvent> eventIter = network.getChaincodeEvents("CHAINCODE_NAME");
        try {
            eventIter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        } finally {
            eventIter.close();
        }

        GatewayRuntimeException e = catchThrowableOfType(
                () -> eventIter.forEachRemaining(event -> { }),
                GatewayRuntimeException.class
        );
        assertThat(e).hasCauseInstanceOf(StatusRuntimeException.class);
        assertThat(e.getStatus().getCode()).isEqualTo(Status.Code.CANCELLED);
    }

    @Test
    void uses_specified_call_options() {
        Deadline expected = Deadline.after(1, TimeUnit.MINUTES);
        CallOption option = CallOption.deadline(expected);
        assertRequestInitiated(() -> network.getChaincodeEvents("CHAINCODE_NAME", option));

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual).first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(expected);
    }

    @Test
    void uses_default_call_options() {
        assertRequestInitiated(() -> network.getChaincodeEvents("CHAINCODE_NAME"));

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual).first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(defaultDeadline);
    }
}
