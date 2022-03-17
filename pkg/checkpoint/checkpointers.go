/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package checkpoint

type Checkpointer interface {
    BlockNumber() uint64
    TransactionID() string
}