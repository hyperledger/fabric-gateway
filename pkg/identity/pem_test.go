/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/stretchr/testify/require"
)

func TestPEM(t *testing.T) {
	privateKey, err := test.NewECDSAPrivateKey()
	require.NoError(t, err)

	certificate, err := test.NewCertificate(privateKey)
	require.NoError(t, err)

	t.Run("Certificate PEM conversion", func(t *testing.T) {
		certificatePEM, err := CertificateToPEM(certificate)
		require.NoError(t, err, "create PEM")

		result, err := CertificateFromPEM(certificatePEM)
		require.NoError(t, err, "create certificate")

		require.True(t, certificate.Equal(result), "Certificates do not match. Expected:\n%v\nGot:\n%v", certificate, result)
	})

	t.Run("Create certificate from PEM fails with invalid PEM", func(t *testing.T) {
		pem := []byte("Non-PEM content")

		_, err := CertificateFromPEM(pem)
		require.Error(t, err)
	})

	t.Run("Create certificate from PEM fails with invalid certificate", func(t *testing.T) {
		pem := []byte("-----BEGIN CERTIFICATE-----\nBAD/DATA-----END CERTIFICATE-----")

		_, err := CertificateFromPEM(pem)
		require.Error(t, err)
	})

	t.Run("Private key PEM conversion", func(t *testing.T) {
		pem, err := PrivateKeyToPEM(privateKey)
		require.NoError(t, err, "create PEM")

		result, err := PrivateKeyFromPEM(pem)
		require.NoError(t, err, "create private key")

		require.True(t, privateKey.Equal(result), "Private keys do not match. Expected:\n%v\nGot:\n%v", privateKey, result)
	})

	t.Run("Create private key fails with invalid PEM", func(t *testing.T) {
		pem := []byte("Non-PEM content")

		_, err := PrivateKeyFromPEM(pem)
		require.Error(t, err)
	})

	t.Run("Create private key from PEM fails with invalid private key data", func(t *testing.T) {
		certificatePEM, err := CertificateToPEM(certificate)
		require.NoError(t, err)

		_, err = PrivateKeyFromPEM([]byte(certificatePEM))
		require.Error(t, err)
	})

	t.Run("Convert bad private key to PEM fails", func(t *testing.T) {
		var privateKey crypto.PrivateKey

		_, err := PrivateKeyToPEM(privateKey)
		require.Error(t, err)
	})
}
