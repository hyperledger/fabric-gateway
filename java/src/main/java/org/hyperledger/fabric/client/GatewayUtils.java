/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import com.google.protobuf.Timestamp;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.protos.msp.Identities;
import org.hyperledger.fabric.protos.orderer.Ab;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.UncheckedIOException;
import java.time.Instant;
import java.util.Arrays;
import java.util.stream.Collectors;

/**
 * Utility functions.
 */
final class GatewayUtils {
    // Private constructor to prevent instantiation
    private GatewayUtils() { }

    public static String toString(final Object o) {
        return o != null ? o.getClass().getSimpleName() + '@' + Integer.toHexString(System.identityHashCode(o)) : "null";
    }

    public static String toString(final Object o, final String... additionalInfo) {
        return toString(o) + Arrays.stream(additionalInfo)
                .collect(Collectors.joining(", ", "(", ")"));
    }

    public static byte[] concat(final byte[]... bytes) {
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

    public static byte[] serializeIdentity(final Identity identity) {
        return Identities.SerializedIdentity.newBuilder()
                .setMspid(identity.getMspId())
                .setIdBytes(ByteString.copyFrom(identity.getCredentials()))
                .build()
                .toByteArray();
    }

    public static void requireNonNullArgument(final Object value, final String message) {
        if (null == value) {
            throw new IllegalArgumentException(message);
        }
    }

    public static Timestamp getCurrentTimestamp() {
        Instant now = Instant.now();
        return Timestamp.newBuilder()
                .setSeconds(now.getEpochSecond())
                .setNanos(now.getNano())
                .build();
    }

    public static Ab.SeekPosition seekLargestBlockNumber() {
        Ab.SeekSpecified largestBlockNumber = Ab.SeekSpecified.newBuilder()
                .setNumber(Long.MAX_VALUE)
                .build();

        return Ab.SeekPosition.newBuilder()
                .setSpecified(largestBlockNumber)
                .build();
    }
}
