/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto"
	"crypto/ecdsa"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
)

func TestPEM(t *testing.T) {
	privateKey, err := test.NewECDSAPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	certificate, err := test.NewCertificate(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	// As in Go 1.15: https://golang.org/pkg/crypto/ecdsa/#PublicKey.Equal
	ecdsaPublicKeyEqual := func(x *ecdsa.PublicKey, y *ecdsa.PublicKey) bool {
		return x.X.Cmp(y.X) == 0 && x.Y.Cmp(y.Y) == 0 && x.Curve == y.Curve
	}

	// As in Go 1.15: https://golang.org/pkg/crypto/ecdsa/#PrivateKey.Equal
	ecdsaPrivateKeyEqual := func(x *ecdsa.PrivateKey, y *ecdsa.PrivateKey) bool {
		return ecdsaPublicKeyEqual(&x.PublicKey, &y.PublicKey) && x.D.Cmp(y.D) == 0
	}

	privateKeyEqual := func(x crypto.PrivateKey, y crypto.PrivateKey) bool {
		switch xx := x.(type) {
		case *ecdsa.PrivateKey:
			yy, ok := y.(*ecdsa.PrivateKey)
			if !ok {
				return false
			}
			return ecdsaPrivateKeyEqual(xx, yy)
		default:
			return false
		}
	}

	t.Run("Certificate PEM conversion", func(t *testing.T) {
		certificatePEM, err := CertificateToPEM(certificate)
		if err != nil {
			t.Fatalf("Failed to create PEM: %v", err)
		}

		result, err := CertificateFromPEM(certificatePEM)
		if err != nil {
			t.Fatalf("Failed to create certificate: %v", err)
		}

		if !certificate.Equal(result) {
			t.Fatalf("Certificates do not match. Expected:\n%v\nGot:\n%v", certificate, result)
		}
	})

	t.Run("Create certificate from PEM fails with invalid PEM", func(t *testing.T) {
		pem := []byte("Non-PEM content")

		_, err := CertificateFromPEM(pem)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("Create certificate from PEM fails with invalid certificate", func(t *testing.T) {
		pem := []byte("-----BEGIN CERTIFICATE-----\nBAD/DATA-----END CERTIFICATE-----")

		_, err := CertificateFromPEM(pem)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("Private key PEM conversion", func(t *testing.T) {
		pem, err := PrivateKeyToPEM(privateKey)
		if err != nil {
			t.Fatalf("Failed to create PEM: %v", err)
		}

		result, err := PrivateKeyFromPEM(pem)
		if err != nil {
			t.Fatalf("Failed to create private key: %v", err)
		}

		if !privateKeyEqual(privateKey, result) {
			t.Fatalf("Keys do not match. Expected:\n%v\nGot:\n%v", privateKey, result)
		}
	})

	t.Run("Create private key fails with invalid PEM", func(t *testing.T) {
		pem := []byte("Non-PEM content")

		_, err := PrivateKeyFromPEM(pem)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("Create private key from PEM fails with invalid private key data", func(t *testing.T) {
		certificatePEM, err := CertificateToPEM(certificate)
		if err != nil {
			t.Fatal(err)
		}

		if _, err = PrivateKeyFromPEM([]byte(certificatePEM)); err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("Convert bad private key to PEM fails", func(t *testing.T) {
		var privateKey crypto.PrivateKey

		_, err := PrivateKeyToPEM(privateKey)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})
}
