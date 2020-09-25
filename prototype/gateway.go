/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"

	gateway "github.com/hyperledger/fabric-gateway/pkg/server"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func main() {
	host := flag.String("h", "peer0.org1.example.com", "hostname of the bootstrap peer")
	port := flag.Int("p", 7051, "port number of the bootstrap peer")
	mspid := flag.String("m", "", "MSP ID of org")
	tlsPath := flag.String("tlscert", "", "path to the org's TLS Certificate")
	idPath := flag.String("id", "", "path to the gateway's wallet identity")
	certPath := flag.String("cert", "", "path to the gateway's Certificate")
	keyPath := flag.String("key", "", "path to the gateway's private key")

	flag.Parse()

	var cert, key string
	// extract bootstrap config from command line flags
	if *idPath != "" {
		id, err := readWalletIdentity(*idPath)
		if err != nil {
			log.Fatalf("failed to read gateway identity: %s", err)
		}
		cert = id.Credentials.Certificate
		key = id.Credentials.Key
	} else {
		f, err := ioutil.ReadFile(*certPath)
		if err != nil {
			log.Fatalf("Failed to read gateway cert: %s", err)
		}
		cert = string(f)
		f, err = ioutil.ReadFile(*keyPath)
		if err != nil {
			log.Fatalf("Failed to read gateway key: %s", err)
		}
		key = string(f)
	}

	pem, err := ioutil.ReadFile(*tlsPath)
	if err != nil {
		log.Fatalf("Failed to read TLS cert: %s", err)
	}

	config := &bootstrapconfig{
		bootstrapPeer: &gateway.PeerEndpoint{
			Host:    *host,
			Port:    uint32(*port),
			TLSCert: pem,
		},
		mspid: *mspid,
		cert:  cert,
		key:   key,
	}

	// setup server and listen
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 1234))
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	gwServer, _ := gateway.NewGatewayServer(config)

	grpcServer := grpc.NewServer()
	pb.RegisterGatewayServer(grpcServer, gwServer)
	//... // determine whether to use TLS
	grpcServer.Serve(lis)
}

type bootstrapconfig struct {
	bootstrapPeer *gateway.PeerEndpoint
	mspid         string
	cert          string
	key           string
}

func (bc *bootstrapconfig) BootstrapPeer() gateway.PeerEndpoint {
	return *bc.bootstrapPeer
}

func (bc *bootstrapconfig) MspID() string {
	return bc.mspid
}

func (bc *bootstrapconfig) Certificate() string {
	return bc.cert
}

func (bc *bootstrapconfig) Key() string {
	return bc.key
}

type x509Identity struct {
	Version     int         `json:"version"`
	MspID       string      `json:"mspId"`
	IDType      string      `json:"type"`
	Credentials credentials `json:"credentials"`
}

type credentials struct {
	Certificate string `json:"certificate"`
	Key         string `json:"privateKey"`
}

// readWalletIdentity loads a user's credentials from a filesystem wallet
func readWalletIdentity(pathname string) (*x509Identity, error) {
	content, err := ioutil.ReadFile(filepath.Clean(pathname))
	if err != nil {
		return nil, err
	}

	id := &x509Identity{}

	if err := json.Unmarshal(content, &id); err != nil {
		return nil, errors.Wrap(err, "Invalid identity format")
	}

	return id, err
}
