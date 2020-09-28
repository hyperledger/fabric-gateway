/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"bytes"
	"testing"
)

func TestHash(t *testing.T) {
	t.Run("Hashes of different data are not identical", func(t *testing.T) {
		foo := []byte("foo")
		bar := []byte("bar")

		fooHash, err := Hash(foo)
		if err != nil {
			t.Fatalf("Failed to hash %s", foo)
		}

		barHash, err := Hash(bar)
		if err != nil {
			t.Fatalf("Failed to has %s", bar)
		}

		if bytes.Equal(fooHash, barHash) {
			t.Fatalf("Hashes for %s and %s were identical: %v", foo, bar, fooHash)
		}
	})
}
