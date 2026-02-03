/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.io.CharArrayWriter;
import java.io.PrintStream;
import java.io.PrintWriter;
import java.util.List;
import java.util.function.Consumer;
import org.hyperledger.fabric.protos.gateway.ErrorDetail;

class GrpcStackTracePrinter {
    private final Consumer<PrintWriter> printStackTraceFn;
    private final GrpcStatus grpcStatus;

    GrpcStackTracePrinter(final Consumer<PrintWriter> printStackTraceFn, final GrpcStatus grpcStatus) {
        this.printStackTraceFn = printStackTraceFn;
        this.grpcStatus = grpcStatus;
    }

    void printStackTrace(final PrintStream out) {
        @SuppressWarnings("PMD.RelianceOnDefaultCharset")
        PrintWriter writer = new PrintWriter(out);
        printStackTrace(writer);
        writer.flush();
    }

    void printStackTrace(final PrintWriter out) {
        CharArrayWriter message = new CharArrayWriter();

        try (PrintWriter printer = new PrintWriter(message)) {
            printStackTraceFn.accept(printer);
        }

        List<ErrorDetail> details = grpcStatus.getDetails();
        if (!details.isEmpty()) {
            message.append("Error details:");
            for (ErrorDetail detail : details) {
                message.append("\n  - address: ")
                        .append(detail.getAddress())
                        .append("\n    mspId: ")
                        .append(detail.getMspId())
                        .append("\n    message: ")
                        .append(detail.getMessage());
            }
            message.append('\n');
        }

        out.print(message);
    }
}
