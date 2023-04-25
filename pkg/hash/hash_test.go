/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package hash

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	for testName, testCase := range map[string]struct {
		hash Hash
	}{
		"NONE":     {hash: NONE},
		"SHA256":   {hash: SHA256},
		"SHA384":   {hash: SHA384},
		"SHA3_256": {hash: SHA3_256},
		"SHA3_384": {hash: SHA3_384},
	} {
		t.Run(testName, func(t *testing.T) {
			t.Run("Hashes of identical data are identical", func(t *testing.T) {
				message := []byte("foobar")

				hash1 := testCase.hash(message)
				hash2 := testCase.hash(message)

				require.EqualValues(t, hash1, hash2)
			})

			t.Run("Hashes of different data are not identical", func(t *testing.T) {
				foo := []byte("foo")
				bar := []byte("bar")

				fooHash := testCase.hash(foo)
				barHash := testCase.hash(bar)

				require.NotEqualValues(t, fooHash, barHash)
			})
		})
	}

	t.Run("NONE returns input", func(t *testing.T) {
		message := []byte("foobar")
		result := NONE(message)
		require.EqualValues(t, message, result)
	})
}
