/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Represents a client identity used to interact with a Fabric network. The identity consists of an identifier for the
 * organization to which the identity belongs, and implementation-specific credentials describing the identity.
 */
export interface Identity {
    /**
     * Member services provider to which this identity is associated.
     */
    mspId: string;

    /**
     * Implementation-specific credentials. For an identity described by a X.509 certificate, the credentials are the
     * PEM-encoded certificate.
     */
    credentials: Uint8Array;
}
