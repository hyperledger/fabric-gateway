/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package hash provides hash implementations used for digital signature of messages sent to a Fabric network.
package hash

import (
	"crypto/sha256"
)

// Hash function generates a digest for the supplied message.
type Hash = func(message []byte) []byte

// SHA256 hash the supplied message bytes to create a digest for signing.
func SHA256(message []byte) []byte {
	digest := sha256.Sum256(message)
	return digest[:]
}
