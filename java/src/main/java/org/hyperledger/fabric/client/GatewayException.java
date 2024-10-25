/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.StatusRuntimeException;
import java.io.PrintStream;
import java.io.PrintWriter;
import java.util.List;
import org.hyperledger.fabric.protos.gateway.ErrorDetail;

/**
 * Thrown if an error is encountered while invoking gRPC services on a gateway peer. Since the gateway delegates much
 * of the processing to other nodes (endorsing peers and orderers), then the error could have originated from one or
 * more of those nodes. In that case, the details will contain errors information from those nodes.
 */
public class GatewayException extends Exception {
    // Ignore similarity with unchecked GatewayRuntimeException - CPD-OFF

    private static final long serialVersionUID = 1L;

    private final transient GrpcStatus grpcStatus;

    /**
     * Constructs a new exception with the specified cause.
     * @param cause the cause.
     */
    public GatewayException(final StatusRuntimeException cause) {
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

    /**
     * {@inheritDoc}
     * This implementation appends any gRPC error details to the stack trace.
     */
    @Override
    public void printStackTrace() {
        printStackTrace(System.err);
    }

    /**
     * {@inheritDoc}
     * This implementation appends any gRPC error details to the stack trace.
     */
    @Override
    public void printStackTrace(final PrintStream out) {
        new GrpcStackTracePrinter(super::printStackTrace, grpcStatus).printStackTrace(out);
    }

    /**
     * {@inheritDoc}
     * This implementation appends any gRPC error details to the stack trace.
     */
    @Override
    public void printStackTrace(final PrintWriter out) {
        new GrpcStackTracePrinter(super::printStackTrace, grpcStatus).printStackTrace(out);
    }

    // CPD-ON
}
