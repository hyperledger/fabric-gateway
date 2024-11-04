/*
 * Copyright 2024 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import static org.assertj.core.api.Assertions.assertThat;

import com.google.protobuf.Any;
import com.google.rpc.Code;
import io.grpc.StatusRuntimeException;
import io.grpc.protobuf.StatusProto;
import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import org.hyperledger.fabric.protos.gateway.ErrorDetail;
import org.junit.jupiter.api.Test;

abstract class CommonGatewayExceptionTest {
    protected abstract Exception newInstance(StatusRuntimeException e);

    @Test
    void error_details_are_printed() {
        List<ErrorDetail> details = Arrays.asList(
                ErrorDetail.newBuilder()
                        .setAddress("ADDRESS1")
                        .setMspId("MSPID1")
                        .setMessage("MESSAGE1")
                        .build(),
                ErrorDetail.newBuilder()
                        .setAddress("ADDRESS2")
                        .setMspId("MSPID2")
                        .setMessage("MESSAGE2")
                        .build());
        Exception e = newInstance(newStatusRuntimeException(Code.ABORTED, "STATUS_MESSAGE", details));

        ByteArrayOutputStream actual = new ByteArrayOutputStream();
        try (PrintStream out = new PrintStream(actual)) {
            e.printStackTrace(out);
        }

        List<String> expected = details.stream()
                .flatMap(detail -> Stream.of(detail.getAddress(), detail.getMspId(), detail.getMessage()))
                .collect(Collectors.toList());
        assertThat(actual.toString()).contains(expected);
    }

    @Test
    void message_from_StatusRuntimeException_is_printed() {
        Exception e = newInstance(newStatusRuntimeException(Code.ABORTED, "STATUS_MESSAGE", Collections.emptyList()));

        ByteArrayOutputStream actual = new ByteArrayOutputStream();
        try (PrintStream out = new PrintStream(actual)) {
            e.printStackTrace(out);
        }

        String expected = e.getCause().getLocalizedMessage();
        assertThat(actual.toString()).contains(expected);
    }

    @Test
    void print_stream_passed_to_printStackTrace_not_closed() {
        Exception e = newInstance(newStatusRuntimeException(Code.ABORTED, "", Collections.emptyList()));

        ByteArrayOutputStream actual = new ByteArrayOutputStream();
        try (PrintStream out = new PrintStream(actual)) {
            e.printStackTrace(out);
            out.println("EXPECTED_SUBSEQUENT_MESSAGE");
        }

        assertThat(actual.toString()).contains("EXPECTED_SUBSEQUENT_MESSAGE");
    }

    private StatusRuntimeException newStatusRuntimeException(Code code, String message, List<ErrorDetail> details) {
        List<Any> anyDetails = details.stream().map(Any::pack).collect(Collectors.toList());
        com.google.rpc.Status status = com.google.rpc.Status.newBuilder()
                .setCode(code.getNumber())
                .setMessage(message)
                .addAllDetails(anyDetails)
                .build();
        return StatusProto.toStatusRuntimeException(status);
    }
}
