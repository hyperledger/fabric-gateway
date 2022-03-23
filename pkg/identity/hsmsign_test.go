//go:build pkcs11
// +build pkcs11

/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func findSoftHSMLibrary(t *testing.T) string {
	libraryLocations := []string{
		"/usr/lib/softhsm/libsofthsm2.so",
		"/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so",
		"/usr/local/lib/softhsm/libsofthsm2.so",
		"/usr/lib/libacsp-pkcs11.so",
	}

	for _, libraryLocation := range libraryLocations {
		if _, err := os.Stat(libraryLocation); !errors.Is(err, os.ErrNotExist) {
			return libraryLocation
		}
	}

	require.FailNow(t, "No SoftHSM library can be found. Tests cannot be run.")
	return ""
}

func TestNewHSMSignerFactory(t *testing.T) {
	t.Run("No library provided", func(t *testing.T) {
		_, err := NewHSMSignerFactory("")
		require.EqualError(t, err, "library path not provided", "Expected Error message about no library provided")
	})

	t.Run("No library found", func(t *testing.T) {
		_, err := NewHSMSignerFactory("/some/path/notalibrary.so")
		require.EqualError(t, err, "instantiation failed for /some/path/notalibrary.so", "Expected Error message about bad library")
	})

	t.Run("Library found", func(t *testing.T) {

		// SoftHSM must be installed for this to pass
		hsmSignerFactory, err := NewHSMSignerFactory(findSoftHSMLibrary(t))
		require.NoError(t, err)
		hsmSignerFactory.Dispose()
	})
}

func TestNewHSMSigner(t *testing.T) {
	// SoftHSM must be installed for this to pass
	hsmSignerFactory, err := NewHSMSignerFactory(findSoftHSMLibrary(t))
	require.NoError(t, err)
	defer hsmSignerFactory.Dispose()

	t.Run("No options provided", func(t *testing.T) {
		hsmSignerOptions := HSMSignerOptions{}
		_, _, err := hsmSignerFactory.NewHSMSigner(hsmSignerOptions)
		require.EqualError(t, err, "no Label provided", "Expected Error message about label")
	})

	t.Run("No Label provided", func(t *testing.T) {
		hsmSignerOptions := HSMSignerOptions{
			Pin:        "12345",
			Identifier: "Fred",
		}
		_, _, err := hsmSignerFactory.NewHSMSigner(hsmSignerOptions)
		require.EqualError(t, err, "no Label provided", "Expected Error message about label")
	})

	t.Run("No Pin provided", func(t *testing.T) {
		hsmSignerOptions := HSMSignerOptions{
			Label:      "12345",
			Identifier: "Fred",
		}
		_, _, err := hsmSignerFactory.NewHSMSigner(hsmSignerOptions)
		require.EqualError(t, err, "no Pin provided", "Expected Error message about pin")
	})

	t.Run("No Identifier provided", func(t *testing.T) {
		hsmSignerOptions := HSMSignerOptions{
			Pin:   "12345",
			Label: "Fred",
		}
		_, _, err := hsmSignerFactory.NewHSMSigner(hsmSignerOptions)
		require.EqualError(t, err, "no Identifier provided", "Expected Error message about identifier")
	})

	t.Run("Label does not exist", func(t *testing.T) {
		hsmSignerOptions := HSMSignerOptions{
			Pin:        "98765432",
			Label:      "NoTokenWithThisLabel",
			Identifier: "fred",
		}
		_, _, err := hsmSignerFactory.NewHSMSigner(hsmSignerOptions)
		require.EqualError(t, err, "could not find token with label NoTokenWithThisLabel", "Expected Error message about label")
	})

	t.Run("Pin Incorrect", func(t *testing.T) {

		// Token with "ForFabric" Label must have been created for this test to pass
		hsmSignerOptions := HSMSignerOptions{
			Pin:        "911",
			Label:      "ForFabric",
			Identifier: "fred",
		}
		_, _, err := hsmSignerFactory.NewHSMSigner(hsmSignerOptions)

		require.Error(t, err)
		require.Contains(t, err.Error(), "CKR_PIN_INCORRECT", "Expected Error message invalid Pin")
	})

	t.Run("Object not found", func(t *testing.T) {

		// Token with "ForFabric" Label must have been created for this test to pass
		hsmSignerOptions := HSMSignerOptions{
			Pin:        "98765432",
			Label:      "ForFabric",
			Identifier: "ThisDoesNotExist",
		}
		_, _, err := hsmSignerFactory.NewHSMSigner(hsmSignerOptions)
		require.EqualError(t, err, "HSM Object not found for key [54686973446f65734e6f744578697374]", "Expected Error message with object not found")
	})
}
