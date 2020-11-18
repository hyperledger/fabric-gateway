/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import java.security.GeneralSecurityException;

/**
 * A signing implementation used to generate digital signatures.
 */
@FunctionalInterface
public interface Signer {
    /**
     * Signs the supplied message digest.
     * @return A digital signature.
     */
    byte[] sign(byte[] digest) throws GeneralSecurityException;
}
