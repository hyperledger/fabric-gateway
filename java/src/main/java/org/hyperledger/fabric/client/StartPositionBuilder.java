/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.orderer.Ab;
import java.util.OptionalLong;
import java.util.Optional;

final class StartPositionBuilder implements Builder<Ab.SeekPosition> {
    private final Ab.SeekPosition.Builder builder = Ab.SeekPosition.newBuilder()
            .setNextCommit(Ab.SeekNextCommit.getDefaultInstance());

    public StartPositionBuilder startBlock(final long blockNumber) {
        Ab.SeekSpecified specified = Ab.SeekSpecified.newBuilder().setNumber(blockNumber).build();
        builder.setSpecified(specified);
        return this;
    }

    public OptionalLong checkpoint(final Checkpoint checkpoint) {
        long blockNumber = checkpoint.getBlockNumber();
        Optional<String> transactionId = checkpoint.getTransactionId();
        if (blockNumber == 0 && !transactionId.isPresent()) {
            return OptionalLong.empty();
        }
        return OptionalLong.of(blockNumber);
    }

    @Override
    public Ab.SeekPosition build() {
        return builder.build();
    }
}
