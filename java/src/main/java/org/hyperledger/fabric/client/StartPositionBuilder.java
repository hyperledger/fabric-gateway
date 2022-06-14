/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.protos.orderer.SeekNextCommit;
import org.hyperledger.fabric.protos.orderer.SeekPosition;
import org.hyperledger.fabric.protos.orderer.SeekSpecified;

final class StartPositionBuilder implements Builder<SeekPosition> {
    private final SeekPosition.Builder builder = SeekPosition.newBuilder()
            .setNextCommit(SeekNextCommit.getDefaultInstance());

    public StartPositionBuilder startBlock(final long blockNumber) {
        SeekSpecified specified = SeekSpecified.newBuilder().setNumber(blockNumber).build();
        builder.setSpecified(specified);
        return this;
    }

    @Override
    public SeekPosition build() {
        return builder.build();
    }
}
