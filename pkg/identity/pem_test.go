/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"strings"
	"testing"
)

func TestPEM(t *testing.T) {
	const certificatePEM = `-----BEGIN CERTIFICATE-----
MIIDujCCAqKgAwIBAgIIE31FZVaPXTUwDQYJKoZIhvcNAQEFBQAwSTELMAkGA1UE
BhMCVVMxEzARBgNVBAoTCkdvb2dsZSBJbmMxJTAjBgNVBAMTHEdvb2dsZSBJbnRl
cm5ldCBBdXRob3JpdHkgRzIwHhcNMTQwMTI5MTMyNzQzWhcNMTQwNTI5MDAwMDAw
WjBpMQswCQYDVQQGEwJVUzETMBEGA1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwN
TW91bnRhaW4gVmlldzETMBEGA1UECgwKR29vZ2xlIEluYzEYMBYGA1UEAwwPbWFp
bC5nb29nbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfRrObuSW5T7q
5CnSEqefEmtH4CCv6+5EckuriNr1CjfVvqzwfAhopXkLrq45EQm8vkmf7W96XJhC
7ZM0dYi1/qOCAU8wggFLMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAa
BgNVHREEEzARgg9tYWlsLmdvb2dsZS5jb20wCwYDVR0PBAQDAgeAMGgGCCsGAQUF
BwEBBFwwWjArBggrBgEFBQcwAoYfaHR0cDovL3BraS5nb29nbGUuY29tL0dJQUcy
LmNydDArBggrBgEFBQcwAYYfaHR0cDovL2NsaWVudHMxLmdvb2dsZS5jb20vb2Nz
cDAdBgNVHQ4EFgQUiJxtimAuTfwb+aUtBn5UYKreKvMwDAYDVR0TAQH/BAIwADAf
BgNVHSMEGDAWgBRK3QYWG7z2aLV29YG2u2IaulqBLzAXBgNVHSAEEDAOMAwGCisG
AQQB1nkCBQEwMAYDVR0fBCkwJzAloCOgIYYfaHR0cDovL3BraS5nb29nbGUuY29t
L0dJQUcyLmNybDANBgkqhkiG9w0BAQUFAAOCAQEAH6RYHxHdcGpMpFE3oxDoFnP+
gtuBCHan2yE2GRbJ2Cw8Lw0MmuKqHlf9RSeYfd3BXeKkj1qO6TVKwCh+0HdZk283
TZZyzmEOyclm3UGFYe82P/iDFt+CeQ3NpmBg+GoaVCuWAARJN/KfglbLyyYygcQq
0SgeDh8dRKUiaW3HQSoYvTvdTuqzwK4CXsr3b5/dAOY8uMuG/IAR3FgwTbZ1dtoW
RvOTa8hYiU6A475WuZKyEHcwnGYe57u2I2KbMgcKjPniocj4QzgYsVAVKW3IwaOh
yE+vPxsiUkvQHdO2fojCkY8jg70jxM+gu59tPDNbw3Uh/2Ij310FgTHsnGQMyA==
-----END CERTIFICATE-----`

	newECDSAPrivateKey := func(t *testing.T) *ecdsa.PrivateKey {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate key")
		}
		return privateKey
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
		certificate, err := CertificateFromPEM([]byte(certificatePEM))
		if err != nil {
			t.Fatalf("Failed to create certificate: %v", err)
		}

		certBytes, err := CertificateToPEM(certificate)
		if err != nil {
			t.Fatalf("Failed to create PEM: %v", err)
		}

		resultPEM := strings.TrimSpace(string(certBytes))
		if certificatePEM != resultPEM {
			t.Fatalf("Input and output PEM does not match. Expected:\n%s\nGot:\n%s", certificatePEM, resultPEM)
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
		inputKey := newECDSAPrivateKey(t)

		pem, err := PrivateKeyToPEM(inputKey)
		if err != nil {
			t.Fatalf("Failed to create PEM: %v", err)
		}

		outputKey, err := PrivateKeyFromPEM(pem)
		if err != nil {
			t.Fatalf("Failed to create private key: %v", err)
		}

		if !privateKeyEqual(inputKey, outputKey) {
			t.Fatalf("Keys do not match. Expected:\n%v\nGot:\n%v", inputKey, outputKey)
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
		_, err := PrivateKeyFromPEM([]byte(certificatePEM))
		if err == nil {
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
