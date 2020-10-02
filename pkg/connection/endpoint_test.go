/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package network

import (
	"bytes"
	"crypto/x509"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
)

func TestEndpoint(t *testing.T) {
	privateKey, err := test.NewECDSAPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	certificate, err := test.NewCertificate(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Creates an empty CertPool for no root certificates", func(t *testing.T) {
		endpoint := Endpoint{}

		certPool := endpoint.tlsRootCAs()

		subjects := certPool.Subjects()
		if len(subjects) != 0 {
			t.Fatalf("Expected no subject, got: %v", subjects)
		}
	})

	t.Run("Creates a CertPool containing root certificates", func(t *testing.T) {
		caCerts := []*x509.Certificate{certificate}
		endpoint := Endpoint{
			TLSRootCertificates: caCerts,
		}

		certPool := endpoint.tlsRootCAs()

		subjects := certPool.Subjects()
		if len(subjects) != len(caCerts) {
			t.Fatalf("Expected %d subjects, got: %v", len(caCerts), subjects)
		}

		for i, actual := range subjects {
			expected := caCerts[i].RawSubject
			if !bytes.Equal(expected, actual) {
				t.Fatalf("Subjects did not match. Expected: %v, got: %v", expected, actual)
			}
		}
	})
}
