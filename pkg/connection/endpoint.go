/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package connection

import (
	"crypto/x509"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Endpoint for a Fabric node, such as a Gateway, Peer or Orderer.
type Endpoint struct {
	Host                string
	Port                uint16
	TLSRootCertificates []*x509.Certificate
}

// Dial creates a gRPC client connection to the endpoint.
func (endpoint *Endpoint) Dial() (*grpc.ClientConn, error) {
	if !endpoint.isTLS() {
		return grpc.Dial(endpoint.String(), grpc.WithInsecure())
	}

	tlsRootCAs := endpoint.tlsRootCAs()
	transportCredentials := credentials.NewClientTLSFromCert(tlsRootCAs, "")
	return grpc.Dial(endpoint.String(), grpc.WithTransportCredentials(transportCredentials))
}

func (endpoint *Endpoint) isTLS() bool {
	return len(endpoint.TLSRootCertificates) > 0
}

// String representation of the endpoint address
func (endpoint *Endpoint) String() string {
	return endpoint.Host + ":" + strconv.FormatUint(uint64(endpoint.Port), 10)
}

func (endpoint *Endpoint) tlsRootCAs() *x509.CertPool {
	certPool := x509.NewCertPool()
	for _, certificate := range endpoint.TLSRootCertificates {
		certPool.AddCert(certificate)
	}
	return certPool
}
