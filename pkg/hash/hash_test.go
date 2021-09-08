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
	t.Run("SHA256", func(t *testing.T) {
		t.Run("Hashes of identical data are identical", func(t *testing.T) {
			message := []byte("foobar")

			hash1 := SHA256(message)
			hash2 := SHA256(message)

			require.EqualValues(t, hash1, hash2)
		})

		t.Run("Hashes of different data are not identical", func(t *testing.T) {
			foo := []byte("foo")
			bar := []byte("bar")

			fooHash := SHA256(foo)
			barHash := SHA256(bar)

			require.NotEqualValues(t, fooHash, barHash)
		})
	})
}
