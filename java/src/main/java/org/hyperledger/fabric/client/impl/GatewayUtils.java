/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.io.UncheckedIOException;
import java.util.Arrays;
import java.util.stream.Collectors;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.protos.msp.Identities;

/**
 * Utility functions.
 */
public final class GatewayUtils {
    // Private constructor to prevent instantiation
    private GatewayUtils() { }

    public static String toString(final Object o) {
        return o != null ? o.getClass().getSimpleName() + '@' + Integer.toHexString(System.identityHashCode(o)) : "null";
    }

    public static String toString(final Object o, final String... additionalInfo) {
        return toString(o) + Arrays.stream(additionalInfo)
                .collect(Collectors.joining(", ", "(", ")"));
    }

    public static void copy(final InputStream input, final OutputStream output) throws IOException {
        for (int b; (b = input.read()) >= 0; ) { // checkstyle:ignore-line:InnerAssignment
            output.write(b);
        }
    }

    public static byte[] concat(byte[]... bytes) {
        int length = Arrays.stream(bytes).mapToInt(b -> b.length).sum();
        try (ByteArrayOutputStream byteOut = new ByteArrayOutputStream(length)) {
            for (byte[] b : bytes) {
                byteOut.write(b);
            }
            return byteOut.toByteArray();
        } catch (IOException e) {
            throw new UncheckedIOException(e);
        }
    }

    public static byte[] serializeIdentity(Identity identity) {
        return Identities.SerializedIdentity.newBuilder()
                .setMspid(identity.getMspId())
                .setIdBytes(ByteString.copyFrom(identity.getCredentials()))
                .build()
                .toByteArray();
    }
}
