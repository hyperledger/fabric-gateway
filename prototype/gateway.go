/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/hyperledger/fabric-gateway/pkg/gateway"
	"github.com/hyperledger/fabric-gateway/pkg/util"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"google.golang.org/grpc"
)

func main() {
	host := flag.String("h", "peer0.org1.example.com", "hostname of the bootstrap peer")
	port := flag.Int("p", 7051, "port number of the bootstrap peer")
	mspid := flag.String("m", "", "MSP ID of org")
	tlsPath := flag.String("tlscert", "", "path to the org's TLS Certificate")
	idPath := flag.String("id", "", "path to the gateway's wallet identity")

	flag.Parse()

	// extract bootstrap config from command line flags
	id, err := util.ReadWalletIdentity(*idPath)
	if err != nil {
		log.Fatalf("failed to read gateway identity: %s", err)
	}

	pem, err := ioutil.ReadFile(*tlsPath)
	if err != nil {
		log.Fatalf("Failed to read TLS cert: %s", err)
	}

	config := &bootstrapconfig{
		bootstrapPeer: &gateway.PeerEndpoint{*host, uint32(*port), pem},
		mspid:         *mspid,
		cert:          id.Credentials.Certificate,
		key:           id.Credentials.Key,
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
