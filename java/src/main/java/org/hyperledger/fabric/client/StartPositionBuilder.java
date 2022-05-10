/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.orderer.Ab;

final class StartPositionBuilder implements Builder<Ab.SeekPosition> {
    private final Ab.SeekPosition.Builder builder = Ab.SeekPosition.newBuilder()
            .setNextCommit(Ab.SeekNextCommit.getDefaultInstance());

    public StartPositionBuilder startBlock(final long blockNumber) {
        Ab.SeekSpecified specified = Ab.SeekSpecified.newBuilder().setNumber(blockNumber).build();
        builder.setSpecified(specified);
        return this;
    }

    @Override
    public Ab.SeekPosition build() {
        return builder.build();
    }
}
