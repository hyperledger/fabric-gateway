/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.List;
import java.util.Objects;
import java.util.stream.Collectors;

import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.protobuf.StatusProto;
import org.hyperledger.fabric.protos.gateway.ErrorDetail;

final class GrpcStatus {
    private final io.grpc.Status status;
    private final io.grpc.Metadata trailers;

    GrpcStatus(final io.grpc.Status status, final io.grpc.Metadata trailers) {
        this.status = status;
        this.trailers = trailers;
    }

    public io.grpc.Status getStatus() {
        return status;
    }

    public List<ErrorDetail> getDetails() {
        return StatusProto.fromStatusAndTrailers(status, trailers)
                .getDetailsList()
                .stream()
                .map(any -> {
                    try {
                        return any.unpack(ErrorDetail.class);
                    } catch (InvalidProtocolBufferException e) {
                        return null;
                    }
                })
                .filter(Objects::nonNull)
                .collect(Collectors.toList());
    }
}
