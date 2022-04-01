/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.TimeUnit;
import java.util.stream.Stream;

import com.google.protobuf.ByteString;
import com.google.protobuf.Option;
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

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.assertj.core.api.Assertions.catchThrowableOfType;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;
class ExpectedRequest {
    String channelName;
    String chaincodeName;
    Ab.SeekPosition.TypeCase typeCase;
    Optional<Long> blockNumber;
    Optional<String> transactionId;

    public ExpectedRequest(String channelName, String chaincodeName, Ab.SeekPosition.TypeCase typeCase, Optional<Long> blockNumber, Optional<String> transactionId){
    this.channelName = channelName;
    this.chaincodeName = chaincodeName;
    this.typeCase = typeCase;
    this.blockNumber = blockNumber;
    this.transactionId = transactionId;
    }

}
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

    void assertChaincodeRequest(org.hyperledger.fabric.protos.gateway.ChaincodeEventsRequest actual, ExpectedRequest expectedRequest){
        assertThat(actual.getChannelId()).isEqualTo(expectedRequest.channelName);
        assertThat(actual.getChaincodeId()).isEqualTo(expectedRequest.chaincodeName);
        assertThat(actual.getStartPosition().getTypeCase()).isEqualTo(expectedRequest.typeCase);
        if(expectedRequest.blockNumber.isPresent()) {
            assertThat(actual.getStartPosition().getSpecified().getNumber()).isEqualTo(expectedRequest.blockNumber.get());
        }
        if(expectedRequest.transactionId.isPresent()) {
            assertThat(actual.getAfterTransactionId()).isEqualTo(expectedRequest.transactionId.get());
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
        try (CloseableIterator<ChaincodeEvent> iter = network.getChaincodeEvents("CHAINCODE_NAME")) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertChaincodeRequest(request, new ExpectedRequest("NETWORK","CHAINCODE_NAME",Ab.SeekPosition.TypeCase.NEXT_COMMIT,Optional.empty(),Optional.empty()));

    }

    @Test
    void sends_valid_request_with_specified_start_block_number() throws Exception {
        long startBlock = 101;
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(startBlock)
                .build();

        try (CloseableIterator<ChaincodeEvent> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertChaincodeRequest(request, new ExpectedRequest("NETWORK","CHAINCODE_NAME",Ab.SeekPosition.TypeCase.SPECIFIED, Optional.ofNullable(startBlock),Optional.empty()));

    }

    @Test
    void sends_valid_request_with_specified_start_block_number_using_sign_bit_for_unsigned_64bit_value() throws Exception {
        long startBlock = -1;
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(startBlock)
                .build();

        try (CloseableIterator<ChaincodeEvent> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertChaincodeRequest(request, new ExpectedRequest("NETWORK","CHAINCODE_NAME",Ab.SeekPosition.TypeCase.SPECIFIED, Optional.ofNullable(startBlock),Optional.empty()));
    }

    @Test
    void uses_specified_start_block_instead_of_unset_checkpoint() throws Exception {
        long startBlock = -1;
        Checkpointer checkpointer = new InMemoryCheckpointer();
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(startBlock)
                .checkpoint(checkpointer)
                .build();

        try (CloseableIterator<ChaincodeEvent> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertChaincodeRequest(request, new ExpectedRequest("NETWORK","CHAINCODE_NAME",Ab.SeekPosition.TypeCase.SPECIFIED, Optional.ofNullable(startBlock),Optional.empty()));
    }

    @Test
    void uses_checkpoint_block_instead_of_specified_start_block() throws Exception {
        long startBlock = -1;
        long blockNumber = 101;
        Checkpointer checkpointer = new InMemoryCheckpointer();
        checkpointer.checkpointBlock(blockNumber);
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(startBlock)
                .checkpoint(checkpointer)
                .build();

        try (CloseableIterator<ChaincodeEvent> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertChaincodeRequest(request, new ExpectedRequest("NETWORK","CHAINCODE_NAME",Ab.SeekPosition.TypeCase.SPECIFIED, Optional.ofNullable(checkpointer.getBlockNumber()),Optional.empty()));
    }

    @Test
    void uses_checkpoint_block_and_transaction_instead_of_specified_start_block() throws Exception {
        long startBlock = -1;
        long blockNumber = 101;
        String transactionId = "TRANSACTION_ID";
        Checkpointer checkpointer = new InMemoryCheckpointer();
        checkpointer.checkpointTransaction(blockNumber, Optional.ofNullable(transactionId));
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(startBlock)
                .checkpoint(checkpointer)
                .build();

        try (CloseableIterator<ChaincodeEvent> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertChaincodeRequest(request, new ExpectedRequest("NETWORK","CHAINCODE_NAME",Ab.SeekPosition.TypeCase.SPECIFIED, Optional.ofNullable(checkpointer.getBlockNumber()),(checkpointer.getTransactionId())));
    }

    @Test
    void start_at_next_commit_with_unset_checkpoint_and_no_start_block() throws Exception {
        Checkpointer checkpointer = new InMemoryCheckpointer();
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .checkpoint(checkpointer)
                .build();

        try (CloseableIterator<ChaincodeEvent> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertChaincodeRequest(request, new ExpectedRequest("NETWORK", "CHAINCODE_NAME", Ab.SeekPosition.TypeCase.NEXT_COMMIT, Optional.empty(),Optional.empty()));
    }

    @Test
    void uses_checkpoint_block_and_transaction_with_unset_start_block() throws Exception {
        long blockNumber = 1;
        String transactionId = "TRANSACTION_ID";
        Checkpointer checkpointer = new InMemoryCheckpointer();
        checkpointer.checkpointTransaction(blockNumber, Optional.ofNullable(transactionId));
        EventsRequest<ChaincodeEvent> eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .checkpoint(checkpointer)
                .build();

        try (CloseableIterator<ChaincodeEvent> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertChaincodeRequest(request, new ExpectedRequest("NETWORK", "CHAINCODE_NAME", Ab.SeekPosition.TypeCase.SPECIFIED, Optional.ofNullable(checkpointer.getBlockNumber()),checkpointer.getTransactionId()));
    }

    @Test
    void sends_valid_request_with_with_start_block_and_checkpointer_chaincode_event() throws  Exception {
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

        try (CloseableIterator<ChaincodeEvent> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertThat(request.getChannelId()).isEqualTo("NETWORK");
        assertThat(request.getChaincodeId()).isEqualTo("CHAINCODE_NAME");

        assertThat(request.getStartPosition().getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.SPECIFIED);
        assertThat(request.getStartPosition().getSpecified().getNumber()).isEqualTo(blockNumber);
        assertThat(request.getAfterTransactionId()).isEqualTo(event.getTxId());
        assertChaincodeRequest(request, new ExpectedRequest("NETWORK", "CHAINCODE_NAME", Ab.SeekPosition.TypeCase.SPECIFIED, Optional.ofNullable(blockNumber), Optional.of(event.getTxId())));

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
        try (CloseableIterator<ChaincodeEvent> iter = network.getChaincodeEvents("CHAINCODE_NAME", option)) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual).first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(expected);
    }

    @Test
    void uses_default_call_options() {
        try (CloseableIterator<ChaincodeEvent> iter = network.getChaincodeEvents("CHAINCODE_NAME")) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        List<CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual).first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(defaultDeadline);
    }
}
