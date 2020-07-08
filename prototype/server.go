/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/gateway"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

var peerURLs = []string{"localhost:7051", "localhost:9051"}

const ordererURL string = "localhost:7050"

func main() {
	flag.Parse()

	// setup server and listen
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 1234))
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	endorserClients, err := connectToPeers()
	if err != nil {
		log.Fatalf("failed to connect to peers: %s", err)
	}

	broadcastClient, err := connectToOrderer()
	if err != nil {
		log.Fatalf("failed to connect to orderer: %s", err)
	}

	gwServer, _ := gateway.NewGatewayServer(ccpPath, endorserClients, broadcastClient)

	grpcServer := grpc.NewServer()
	pb.RegisterGatewayServer(grpcServer, gwServer)
	//... // determine whether to use TLS
	grpcServer.Serve(lis)
}

func connectToPeers() ([]peer.EndorserClient, error) {
	var endorserClients []peer.EndorserClient
	// make client connection to peer
	// organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem
	pemPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"tlsca",
		"tlsca.org1.example.com-cert.pem",
	)
	peerCreds, err := credentials.NewClientTLSFromFile(pemPath, "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read credentials: ")
	}
	peerConn, err := grpc.Dial(peerURLs[0], grpc.WithTransportCredentials(peerCreds))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to peer: ")
	}
	endorserClients = append(endorserClients, peer.NewEndorserClient(peerConn))

	pemPath2 := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org2.example.com",
		"tlsca",
		"tlsca.org2.example.com-cert.pem",
	)
	peerCreds2, err := credentials.NewClientTLSFromFile(pemPath2, "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read credentials: ")
	}
	peerConn2, err := grpc.Dial(peerURLs[1], grpc.WithTransportCredentials(peerCreds2))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to peer: ")
	}
	endorserClients = append(endorserClients, peer.NewEndorserClient(peerConn2))

	return endorserClients, nil
}

func connectToOrderer() (orderer.AtomicBroadcast_BroadcastClient, error) {
	// make client connection to orderer
	pemPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"ordererOrganizations",
		"example.com",
		"msp",
		"tlscacerts",
		"tlsca.example.com-cert.pem",
	)
	ordererCreds, err := credentials.NewClientTLSFromFile(pemPath, "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read credentials: ")
	}

	kap := keepalive.ClientParameters{
		Time:                300 * time.Second,
		Timeout:             20 * time.Second,
		PermitWithoutStream: true,
	}

	ordererConn, err := grpc.Dial(ordererURL, grpc.WithTransportCredentials(ordererCreds), grpc.WithKeepaliveParams(kap))

	broadcastClient, err := orderer.NewAtomicBroadcastClient(ordererConn).Broadcast(context.TODO())
	if err != nil {
		rpcStatus, ok := status.FromError(err)
		if ok {
			fmt.Println(rpcStatus.Message())
		}
		return nil, errors.Wrap(err, "failed to connect to orderer: ")
	}

	return broadcastClient, nil
}
