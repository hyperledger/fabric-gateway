/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.gateway.CommitStatusResponse;
import org.hyperledger.fabric.protos.peer.TxValidationCode;

/**
 * Information about a transaction that is committed to the ledger.
 */
final class StatusImpl implements Status {
    private final String transactionId;
    private final long blockNumber;
    private final TxValidationCode code;

    StatusImpl(final String transactionId, final CommitStatusResponse response) {
        this.transactionId = transactionId;
        this.blockNumber = response.getBlockNumber();
        this.code = response.getResult();
    }

    @Override
    public String getTransactionId() {
        return transactionId;
    }

    @Override
    public long getBlockNumber() {
        return blockNumber;
    }

    @Override
    public TxValidationCode getCode() {
        return code;
    }

    @Override
    public boolean isSuccessful() {
        return code == TxValidationCode.VALID;
    }

    @Override
    public String toString() {
        return GatewayUtils.toString(this,
                "transactionId: " + transactionId,
                "code: " + code.getNumber() + " (" + code.name() + ")",
                "blockNumber: " + blockNumber);
    }
}
