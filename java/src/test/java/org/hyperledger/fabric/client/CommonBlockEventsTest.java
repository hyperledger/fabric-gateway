/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;
import java.util.List;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.CallOptions;
import io.grpc.Deadline;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.msp.Identities;
import org.hyperledger.fabric.protos.orderer.Ab;
import org.hyperledger.fabric.protos.peer.EventsPackage;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.assertj.core.api.Assertions.catchThrowableOfType;

public abstract class CommonBlockEventsTest<E> {
    private static final Deadline defaultDeadline = Deadline.after(1, TimeUnit.DAYS);

    protected GatewayMocker mocker;
    protected DeliverServiceStub stub;
    protected Network network;

    private Gateway gateway;

    @BeforeEach
    void beforeEach() {
        mocker = new GatewayMocker();
        stub = mocker.getDeliverServiceStubSpy();

        Gateway.Builder builder = mocker.getGatewayBuilder();
        setEventsOptions(builder, CallOption.deadline(defaultDeadline));
        gateway = builder.connect();
        network = gateway.getNetwork("NETWORK");
    }

    @AfterEach
    void afterEach() {
        gateway.close();
        mocker.close();
    }

    protected abstract void setEventsOptions(Gateway.Builder builder, CallOption... options);
    protected abstract EventsPackage.DeliverResponse newDeliverResponse(long blockNumber);
    protected abstract void stubDoThrow(Throwable... t);
    protected abstract CloseableIterator<E> getEvents(CallOption... options);
    protected abstract Stream<Common.Envelope> captureEvents();
    protected abstract EventsBuilder<E> newEventsRequest();
    protected abstract void stubDoReturn(Stream<EventsPackage.DeliverResponse> responses);
    protected abstract E extractEvent(EventsPackage.DeliverResponse response);

    private void assertValidBlockEventsRequestHeader(final Common.Payload payload) throws InvalidProtocolBufferException, CertificateException {
        Common.Header header = payload.getHeader();
        Common.ChannelHeader channelHeader = Common.ChannelHeader.parseFrom(header.getChannelHeader());
        Common.SignatureHeader signatureHeader = Common.SignatureHeader.parseFrom(header.getSignatureHeader());
        Identities.SerializedIdentity creator = Identities.SerializedIdentity.parseFrom(signatureHeader.getCreator());

        String credentials = creator.getIdBytes().toStringUtf8();
        X509Certificate certificate = org.hyperledger.fabric.client.identity.Identities.readX509Certificate(credentials);
        Identity actualIdentity = new X509Identity(creator.getMspid(), certificate);

        assertThat(channelHeader.getChannelId()).isEqualTo(network.getName());
        assertThat(actualIdentity).isEqualTo(gateway.getIdentity());
    }

    private void assertStartPositionSpecified (final Ab.SeekInfo seekInfo, final long startBlock) {
        Ab.SeekPosition start = seekInfo.getStart();
        assertThat(start.getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.SPECIFIED);
        assertThat(start.getSpecified().getNumber()).isEqualTo(startBlock);
 }
    private void assertStartPositionNextCommit (final Ab.SeekInfo seekInfo) {
        Ab.SeekPosition start = seekInfo.getStart();
        assertThat(start.getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.NEXT_COMMIT);
    }

    private void assertStopPosition(final Ab.SeekInfo seekInfo) {
        Ab.SeekPosition stop = seekInfo.getStop();
        assertThat(stop.getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.SPECIFIED);
        assertThat(stop.getSpecified().getNumber()).isEqualTo(Long.MAX_VALUE);
    }

    @Test
    void throws_on_connection_error() {
        StatusRuntimeException expected = new StatusRuntimeException(Status.UNAVAILABLE);
        stubDoThrow(expected);

        GatewayRuntimeException e = catchThrowableOfType(() -> {
            try (CloseableIterator<?> iter = getEvents()) {
                iter.forEachRemaining(event -> { });
            }
        }, GatewayRuntimeException.class);

        assertThat(e.getStatus()).isEqualTo(expected.getStatus());
        assertThat(e).hasCauseInstanceOf(StatusRuntimeException.class);
    }

    @Test
    void sends_valid_request_with_default_start_position() throws Exception {
        try (CloseableIterator<?> iter = getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        Common.Envelope request = captureEvents().findFirst().get();
        Common.Payload payload = Common.Payload.parseFrom(request.getPayload());
        assertValidBlockEventsRequestHeader(payload);

        Ab.SeekInfo seekInfo = Ab.SeekInfo.parseFrom(payload.getData());
        Ab.SeekPosition start = seekInfo.getStart();
        assertThat(start.getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.NEXT_COMMIT);
        Ab.SeekPosition stop = seekInfo.getStop();
        assertThat(stop.getTypeCase()).isEqualTo(Ab.SeekPosition.TypeCase.SPECIFIED);
        assertThat(stop.getSpecified().getNumber()).isEqualTo(Long.MAX_VALUE);
    }

    @Test
    void sends_valid_request_with_specified_start_block_number() throws Exception {
        long startBlock = 101;
        EventsRequest<?> eventsRequest = newEventsRequest()
                .startBlock(startBlock)
                .build();

        try (CloseableIterator<?> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        Common.Envelope request = captureEvents().findFirst().get();
        Common.Payload payload = Common.Payload.parseFrom(request.getPayload());
        assertValidBlockEventsRequestHeader(payload);

        Ab.SeekInfo seekInfo = Ab.SeekInfo.parseFrom(payload.getData());
        assertStartPositionSpecified(seekInfo, startBlock);
        assertStopPosition(seekInfo);
    }

    @Test
    void sends_valid_request_with_specified_start_block_number_using_sign_bit_for_unsigned_64bit_value() throws Exception {
        long startBlock = -1;
        EventsRequest<?> eventsRequest = newEventsRequest()
                .startBlock(startBlock)
                .build();

        try (CloseableIterator<?> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        Common.Envelope request = captureEvents().findFirst().get();
        Common.Payload payload = Common.Payload.parseFrom(request.getPayload());
        assertValidBlockEventsRequestHeader(payload);

        Ab.SeekInfo seekInfo = Ab.SeekInfo.parseFrom(payload.getData());
        assertStartPositionSpecified(seekInfo, startBlock);
        assertStopPosition(seekInfo);
    }

    @Test
    void uses_specified_start_block_instead_of_unset_checkpoint() throws Exception {
        long startBlock = -1;
        Checkpointer checkpointer = new InMemoryCheckpointer();
        EventsRequest<?> eventsRequest = newEventsRequest()
                .startBlock(startBlock)
                .checkpoint(checkpointer)
                .build();

        try (CloseableIterator<?> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        Common.Envelope request = captureEvents().findFirst().get();
        Common.Payload payload = Common.Payload.parseFrom(request.getPayload());
        assertValidBlockEventsRequestHeader(payload);

        Ab.SeekInfo seekInfo = Ab.SeekInfo.parseFrom(payload.getData());
        assertStartPositionSpecified(seekInfo, startBlock);
        assertStopPosition(seekInfo);
    }

    @Test
    void uses_checkpoint_block_instead_of_specified_start_block() throws Exception {
        long blockNumber = 111;
        Checkpointer checkpointer = new InMemoryCheckpointer();
        checkpointer.checkpointBlock(blockNumber);
        EventsRequest<?> eventsRequest = newEventsRequest()
                .startBlock(-1)
                .checkpoint(checkpointer)
                .build();

        try (CloseableIterator<?> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        Common.Envelope request = captureEvents().findFirst().get();
        Common.Payload payload = Common.Payload.parseFrom(request.getPayload());
        assertValidBlockEventsRequestHeader(payload);

        Ab.SeekInfo seekInfo = Ab.SeekInfo.parseFrom(payload.getData());
        assertStartPositionSpecified(seekInfo, blockNumber+1);
        assertStopPosition(seekInfo);
    }

    @Test
    void start_at_next_commit_with_unset_checkpoint_and_no_start_block() throws Exception {
        Checkpointer checkpointer = new InMemoryCheckpointer();
        EventsRequest<?> eventsRequest = newEventsRequest()
                .checkpoint(checkpointer)
                .build();

        try (CloseableIterator<?> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        Common.Envelope request = captureEvents().findFirst().get();
        Common.Payload payload = Common.Payload.parseFrom(request.getPayload());
        assertValidBlockEventsRequestHeader(payload);

        Ab.SeekInfo seekInfo = Ab.SeekInfo.parseFrom(payload.getData());
        assertStartPositionNextCommit(seekInfo);
        assertStopPosition(seekInfo);
    }

    @Test
    void uses_checkpoint_block_zero_with_set_transaction_id_instead_of_specified_start_block() throws Exception {
        long blockNumber = 0;
        Checkpointer checkpointer = new InMemoryCheckpointer();
        checkpointer.checkpointTransaction(blockNumber, "transactionId");
        EventsRequest<?> eventsRequest = newEventsRequest()
                .startBlock(-1)
                .checkpoint(checkpointer)
                .build();

        try (CloseableIterator<?> iter = eventsRequest.getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        Common.Envelope request = captureEvents().findFirst().get();
        Common.Payload payload = Common.Payload.parseFrom(request.getPayload());
        assertValidBlockEventsRequestHeader(payload);

        Ab.SeekInfo seekInfo = Ab.SeekInfo.parseFrom(payload.getData());
        assertStartPositionSpecified(seekInfo, blockNumber);
        assertStopPosition(seekInfo);
    }

    @Test()
    void throws_on_receive_of_status_message() {
        EventsPackage.DeliverResponse response = EventsPackage.DeliverResponse.newBuilder()
                .setStatus(Common.Status.SERVICE_UNAVAILABLE)
                .build();
        stubDoReturn(Stream.of(response));

        assertThatThrownBy(() -> {
            try (CloseableIterator<?> iter = getEvents()) {
                iter.forEachRemaining(event -> { });
            }
        }).hasMessageContaining(Common.Status.SERVICE_UNAVAILABLE.toString());
    }

    @Test
    void returns_events() {
        List<EventsPackage.DeliverResponse> responses = Stream.of(newDeliverResponse(1), newDeliverResponse(2))
                .collect(Collectors.toList());
        List<E> expected = responses.stream()
                .map(this::extractEvent)
                .collect(Collectors.toList());
        stubDoReturn(responses.stream());

        try (CloseableIterator<E> actual = getEvents()) {
            assertThat(actual)
                    .toIterable()
                    .hasSameElementsAs(expected);
        }
    }

    @Test
    void close_stops_receiving_events() {
        Stream<EventsPackage.DeliverResponse> responses = Stream.generate(() -> newDeliverResponse(1)).limit(100);
        stubDoReturn(responses);

        CloseableIterator<?> eventIter = getEvents();
        try {
            eventIter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        } finally {
            eventIter.close();
        }

        // Some events may be buffered at the client end but the number of events should be limited after close
        assertThat(eventIter)
                .toIterable()
                .hasSizeLessThan(100);
    }

    @Test
    void eventing_can_be_restarted_after_close() {
        List<EventsPackage.DeliverResponse> responses = Stream.of(newDeliverResponse(1), newDeliverResponse(2))
                .collect(Collectors.toList());
        List<E> expected = responses.stream()
                .map(this::extractEvent)
                .collect(Collectors.toList());
        stubDoReturn(responses.stream());

        try (CloseableIterator<?> eventIter = getEvents()) {
            eventIter.hasNext(); // Interact with iterator before closing to ensure async request has been made
        }

        stubDoReturn(responses.stream());

        try (CloseableIterator<E> actual = getEvents()) {
            assertThat(actual)
                    .toIterable()
                    .hasSameElementsAs(expected);
        }
    }

    @Test
    void uses_specified_call_options() {
        Deadline expected = Deadline.after(1, TimeUnit.MINUTES);
        CallOption option = CallOption.deadline(expected);
        try (CloseableIterator<?> iter = getEvents(option)) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        List<io.grpc.CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual).first()
                .extracting(io.grpc.CallOptions::getDeadline)
                .isEqualTo(expected);
    }

    @Test
    void uses_default_call_options() {
        try (CloseableIterator<?> iter = getEvents()) {
            iter.hasNext(); // Interact with iterator before asserting to ensure async request has been made
        }

        List<io.grpc.CallOptions> actual = mocker.captureCallOptions();
        assertThat(actual).first()
                .extracting(CallOptions::getDeadline)
                .isEqualTo(defaultDeadline);
    }
}
