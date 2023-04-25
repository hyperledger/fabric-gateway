/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.bouncycastle.jce.provider.BouncyCastleProvider;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.EnumSource;

import java.nio.charset.StandardCharsets;
import java.security.Security;

import static org.assertj.core.api.Assertions.assertThat;

public final class HashTest {
    @BeforeAll
    static void beforeAll() {
        // Required from some Java 1.8 distributions that don't include SHA3 message digests
        Security.addProvider(new BouncyCastleProvider());
    }

    @AfterAll
    static void afterAll() {
        Security.removeProvider(BouncyCastleProvider.PROVIDER_NAME);
    }

    @ParameterizedTest
    @EnumSource(Hash.class)
    void identical_messages_have_identical_hash(Hash hash) {
        byte[] message = "MESSAGE".getBytes(StandardCharsets.UTF_8);

        byte[] hash1 = hash.apply(message);
        byte[] hash2 = hash.apply(message);

        assertThat(hash1).isEqualTo(hash2);
    }

    @ParameterizedTest
    @EnumSource(Hash.class)
    void different_messages_have_different_hash(Hash hash) {
        byte[] foo = "foo".getBytes(StandardCharsets.UTF_8);
        byte[] bar = "bar".getBytes(StandardCharsets.UTF_8);

        byte[] fooHash = hash.apply(foo);
        byte[] barHash = hash.apply(bar);

        assertThat(fooHash).isNotEqualTo(barHash);
    }

    @Test
    void NONE_returns_input() {
        byte[] message = "MESSAGE".getBytes(StandardCharsets.UTF_8);
        byte[] result = Hash.NONE.apply(message);
        assertThat(result).isEqualTo(message);
    }
}
