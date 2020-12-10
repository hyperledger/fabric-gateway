/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Transaction } from './transaction';

export interface Proposal {
    /**
     * Get the serialized proposal message.
     */
    getBytes(): Uint8Array;

    /**
     * Get the digest of the proposal. This is used to generate a digital signature.
     */
    getDigest(): Uint8Array;

    /**
     * Get the transaction ID for this proposal.
     */
    getTransactionId(): string;

    /**
     * Evaluate the transaction proposal and obtain its result, without updating the ledger. This runs the transaction
     * on a peer to obtain a transaction result, but does not submit the endorsed transaction to the orderer to be
     * committed to the ledger.
     */
    evaluate(): Promise<Uint8Array>;

    /**
     * Obtain endorsement for the transaction proposal from sufficient peers to allow it to be committed to the ledger.
     */
    endorse(): Promise<Transaction>;
}
