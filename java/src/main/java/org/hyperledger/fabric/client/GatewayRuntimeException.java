/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.List;

import io.grpc.StatusRuntimeException;
import org.hyperledger.fabric.protos.gateway.ErrorDetail;

/**
 * Thrown if an error is encountered while invoking gRPC services on a gateway peer. Since the gateway delegates much
 * of the processing to other nodes (endorsing peers and orderers), then the error could have originated from one or
 * more of those nodes. In that case, the details will contain errors information from those nodes.
 */
public class GatewayRuntimeException extends RuntimeException {
    private static final long serialVersionUID = 1L;

    private final transient GrpcStatus grpcStatus;

    /**
     * Constructs a new exception with the specified cause.
     * @param cause the cause.
     */
    public GatewayRuntimeException(final StatusRuntimeException cause) {
        super(cause);
        grpcStatus = new GrpcStatus(cause.getStatus(), cause.getTrailers());
    }

    /**
     * Returns the status code as a gRPC Status object.
     * @return gRPC call status.
     */
    public io.grpc.Status getStatus() {
        return grpcStatus.getStatus();
    }

    /**
     * Get the gRPC error details returned by a gRPC invocation failure.
     * @return error details.
     */
    public List<ErrorDetail> getDetails() {
        return grpcStatus.getDetails();
    }
}
