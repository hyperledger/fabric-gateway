/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.Any;
import com.google.rpc.Code;
import io.grpc.StatusRuntimeException;
import io.grpc.protobuf.StatusProto;
import org.hyperledger.fabric.protos.gateway.ErrorDetail;
import org.junit.jupiter.api.Test;

import java.io.CharArrayWriter;
import java.io.PrintWriter;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.assertThat;

public final class GatewayExceptionTest {
    @Test
    void error_details_are_printed() {
        List<ErrorDetail> details = Arrays.asList(
                ErrorDetail.newBuilder().setAddress("ADDRESS1").setMspId("MSPID1").setMessage("MESSAGE1").build(),
                ErrorDetail.newBuilder().setAddress("ADDRESS2").setMspId("MSPID2").setMessage("MESSAGE2").build()
        );
        GatewayException e = new GatewayException(newStatusRuntimeException(Code.ABORTED, "STATUS_MESSAGE", details));

        CharArrayWriter actual = new CharArrayWriter();
        try (PrintWriter out = new PrintWriter(actual)) {
            e.printStackTrace(out);
            out.flush();
        }

        List<String> expected = details.stream()
                .flatMap(detail -> Stream.of(detail.getAddress(), detail.getMspId(), detail.getMessage()))
                .collect(Collectors.toList());
        assertThat(actual.toString()).contains(expected);
    }

    private StatusRuntimeException newStatusRuntimeException(Code code, String message, List<ErrorDetail> details) {
        List<Any> anyDetails = details.stream()
                .map(Any::pack)
                .collect(Collectors.toList());
        com.google.rpc.Status status = com.google.rpc.Status.newBuilder()
                .setCode(code.getNumber())
                .setMessage(message)
                .addAllDetails(anyDetails)
                .build();
        return StatusProto.toStatusRuntimeException(status);
    }
}
