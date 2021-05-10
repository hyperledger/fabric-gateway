/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package hash

import (
	"bytes"
	"testing"
)

func TestHash(t *testing.T) {
	t.Run("SHA256", func(t *testing.T) {
		t.Run("Hashes of identical data are identical", func(t *testing.T) {
			message := []byte("foobar")

			hash1 := SHA256(message)
			hash2 := SHA256(message)

			if !bytes.Equal(hash1, hash2) {
				t.Fatalf("Hashes of %s were not identical:\n%v\n%v", message, hash1, hash2)
			}
		})

		t.Run("Hashes of different data are not identical", func(t *testing.T) {
			foo := []byte("foo")
			bar := []byte("bar")

			fooHash := SHA256(foo)
			barHash := SHA256(bar)

			if bytes.Equal(fooHash, barHash) {
				t.Fatalf("Hashes of %s and %s were identical: %v", foo, bar, fooHash)
			}
		})
	})
}
