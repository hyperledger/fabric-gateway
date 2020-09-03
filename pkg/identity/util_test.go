/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"bytes"
	"testing"
)

func TestUtil(t *testing.T) {
	certificatePEM := `-----BEGIN CERTIFICATE-----
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

	t.Run("Serialize", func(t *testing.T) {
		inputIdentity := &X509Identity{
			mspID:       "mspID",
			certificate: []byte(certificatePEM),
		}

		identityMessage, err := Serialize(inputIdentity)
		if err != nil {
			t.Fatalf("Failed to serialize identity: %v", err)
		}

		outputIdentity, err := Deserialize(identityMessage)
		if err != nil {
			t.Fatalf("Failed to deserialize identity: %v", err)
		}

		if outputIdentity.MspID() != inputIdentity.MspID() {
			t.Fatalf("Expected MspID %s, got %s", inputIdentity.MspID(), outputIdentity.MspID())
		}

		if !bytes.Equal(inputIdentity.Credentials(), outputIdentity.Credentials()) {
			t.Fatalf("Expected Credentials:\n%v\nGot:\n%v", inputIdentity.Credentials(), outputIdentity.Credentials())
		}
	})

	t.Run("Hash", func(t *testing.T) {
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
