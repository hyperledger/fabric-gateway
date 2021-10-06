/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;

final class SubmittedTransactionImpl extends CommitImpl implements SubmittedTransaction {
    private final byte[] result;

    SubmittedTransactionImpl(final GatewayClient client, final SigningIdentity signingIdentity,
                             final String transactionId, final SignedCommitStatusRequest signedRequest,
                             final byte[] result) {
        super(client, signingIdentity, transactionId, signedRequest);
        this.result = result;
    }

    @Override
    public byte[] getResult() {
        return result;
    }
}
