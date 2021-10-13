/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Arrays;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Stream;

import com.google.protobuf.ByteString;
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
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;

public final class ChaincodeEventsTest {
    private static final TestUtils utils = TestUtils.getInstance();

    private GatewayMocker mocker;
    private GatewayServiceStub stub;
    private Gateway gateway;
    private Network network;

    @BeforeEach
    void beforeEach() {
        mocker = new GatewayMocker();
        stub = mocker.getServiceStubSpy();

        gateway = mocker.getGatewayBuilder().connect();
        network = gateway.getNetwork("NETWORK");
    }

    @AfterEach
    void afterEach() {
        gateway.close();
        mocker.close();
    }

    @Test
    void throws_NullPointerException_on_null_CHAINCODE_NAME() {
        assertThatThrownBy(() -> network.getChaincodeEvents(null))
                .isInstanceOf(NullPointerException.class)
                .hasMessageContaining("chaincode name");
    }

    @Test
    void throws_on_connection_error() {
        doThrow(new StatusRuntimeException(Status.UNAVAILABLE)).when(stub).chaincodeEvents(any());

        assertThatThrownBy(() -> {
            try (CloseableIterator<ChaincodeEvent> events = network.getChaincodeEvents("CHAINCODE_NAME")) {
                events.forEachRemaining(event -> {});
            }
        }).isInstanceOf(StatusRuntimeException.class);
    }

    @Test
    void sends_valid_request_with_default_start_position() throws Exception {
        try (CloseableIterator<ChaincodeEvent> iter = network.getChaincodeEvents("CHAINCODE_NAME")) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> {});
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertThat(request.getChannelId()).isEqualTo("NETWORK");
        assertThat(request.getChaincodeId()).isEqualTo("CHAINCODE_NAME");

        assertThat(request.getStartPosition().getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.NEXT_COMMIT);
    }

    @Test
    void sends_valid_request_with_specified_start_block_number() throws Exception {
        long startBlock = 101;
        org.hyperledger.fabric.client.ChaincodeEventsRequest eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(startBlock)
                .build();

        try (CloseableIterator<ChaincodeEvent> iter = eventsRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> {});
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertThat(request.getChannelId()).isEqualTo("NETWORK");
        assertThat(request.getChaincodeId()).isEqualTo("CHAINCODE_NAME");

        assertThat(request.getStartPosition().getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.SPECIFIED);
        assertThat(request.getStartPosition().getSpecified().getNumber()).isEqualTo(startBlock);
    }

    @Test
    void sends_valid_request_with_specified_start_block_number_using_sign_bit_for_unsigned_64bit_value() throws Exception {
        long startBlock = -1;
        org.hyperledger.fabric.client.ChaincodeEventsRequest eventsRequest = network.newChaincodeEventsRequest("CHAINCODE_NAME")
                .startBlock(startBlock)
                .build();

        try (CloseableIterator<ChaincodeEvent> iter = eventsRequest.getEvents()) {
            // Need to interact with iterator before asserting to ensure async request has been made
            iter.forEachRemaining(event -> {});
        }

        SignedChaincodeEventsRequest signedRequest = mocker.captureChaincodeEvents();
        ChaincodeEventsRequest request = ChaincodeEventsRequest.parseFrom(signedRequest.getRequest());

        assertThat(request.getChannelId()).isEqualTo("NETWORK");
        assertThat(request.getChaincodeId()).isEqualTo("CHAINCODE_NAME");

        assertThat(request.getStartPosition().getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.SPECIFIED);
        assertThat(request.getStartPosition().getSpecified().getNumber()).isEqualTo(startBlock);
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
            assertThat(Stream.generate(actual::next).limit(3)).hasSameElementsAs(expected);
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
        ChaincodeEventPackage.ChaincodeEvent event2 = ChaincodeEventPackage.ChaincodeEvent.newBuilder()
                .setChaincodeId("CHAINCODE_NAME")
                .setTxId("tx2")
                .setEventName("event2")
                .setPayload(ByteString.copyFromUtf8("payload2"))
                .build();

        Stream<ChaincodeEventsResponse> responses = Stream.of(
                ChaincodeEventsResponse.newBuilder()
                        .setBlockNumber(1)
                        .addEvents(event1)
                        .build(),
                ChaincodeEventsResponse.newBuilder()
                        .setBlockNumber(2)
                        .addEvents(event2)
                        .build()
        );
        doReturn(responses).when(stub).chaincodeEvents(any());

        CloseableIterator<ChaincodeEvent> eventIter = network.getChaincodeEvents("CHAINCODE_NAME");
        ChaincodeEvent event = eventIter.next();
        assertThat(event).isEqualTo(new ChaincodeEventImpl(1, event1));

        eventIter.close();

        assertThatThrownBy(() -> eventIter.next())
                .isInstanceOf(NoSuchElementException.class);
    }
}
