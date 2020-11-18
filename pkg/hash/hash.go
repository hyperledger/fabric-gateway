/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package hash

import (
	"crypto/sha256"
)

// Hash function generates a digest for the supplied message.
type Hash = func(message []byte) ([]byte, error)

// SHA256 hash the supplied message bytes to create a digest for signing.
func SHA256(message []byte) ([]byte, error) {
	hash := sha256.New()

	_, err := hash.Write(message)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}
