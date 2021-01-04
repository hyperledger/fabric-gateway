/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.nio.charset.StandardCharsets;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

public final class HashTest {
    @Test
    void identical_messages_have_identical_hash() {
        byte[] message = "MESSAGE".getBytes(StandardCharsets.UTF_8);

        byte[] hash1 = Hash.sha256(message);
        byte[] hash2 = Hash.sha256(message);

        assertThat(hash1).isEqualTo(hash2);
    }

    @Test
    void different_messages_have_different_hash() {
        byte[] foo = "foo".getBytes(StandardCharsets.UTF_8);
        byte[] bar = "bar".getBytes(StandardCharsets.UTF_8);

        byte[] fooHash = Hash.sha256(foo);
        byte[] barHash = Hash.sha256(bar);

        assertThat(fooHash).isNotEqualTo(barHash);
    }
}
