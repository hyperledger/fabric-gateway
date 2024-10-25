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

    public GrpcStackTracePrinter(final Consumer<PrintWriter> printStackTraceFn, final GrpcStatus grpcStatus) {
        this.printStackTraceFn = printStackTraceFn;
        this.grpcStatus = grpcStatus;
    }

    public void printStackTrace(final PrintStream out) {
        PrintWriter writer = new PrintWriter(out);
        printStackTrace(writer);
        writer.flush();
    }

    public void printStackTrace(final PrintWriter out) {
        CharArrayWriter message = new CharArrayWriter();

        try (PrintWriter printer = new PrintWriter(message)) {
            printStackTraceFn.accept(printer);
        }

        List<ErrorDetail> details = grpcStatus.getDetails();
        if (!details.isEmpty()) {
            message.append("Error details:\n");
            for (ErrorDetail detail : details) {
                message.append("    address: ")
                        .append(detail.getAddress())
                        .append("; mspId: ")
                        .append(detail.getMspId())
                        .append("; message: ")
                        .append(detail.getMessage())
                        .append('\n');
            }
        }

        out.print(message);
    }
}
