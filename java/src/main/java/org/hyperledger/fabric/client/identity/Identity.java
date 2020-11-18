/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

/**
 * Represents a client identity used to interact with a Fabric network.
 */
public interface Identity {
    /**
     * Member services provider to which this identity is associated.
     * @return A member services provider identifier.
     */
    String getMspId();

    /**
     * Implementation-specific credentials.
     * @return Credential data.
     */
    byte[] getCredentials();
}
