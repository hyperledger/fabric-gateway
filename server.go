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

	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type gatewayServer struct {
	ccpPath string
}

func (gs *gatewayServer) SubmitTransaction(ctx context.Context, txn *pb.Transaction) (*pb.Response, error) {
	contract, err := gs.getContract(txn)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to submit transaction: ")
	}

	result, err := contract.SubmitTransaction(txn.TxnName, txn.Args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to submit transaction: ")
	}

	return &pb.Response{Value: result}, nil
}

func (gs *gatewayServer) EvaluateTransaction(ctx context.Context, txn *pb.Transaction) (*pb.Response, error) {
	contract, err := gs.getContract(txn)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to evaluate transaction: ")
	}

	result, err := contract.EvaluateTransaction(txn.TxnName, txn.Args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to evaluate transaction: ")
	}

	return &pb.Response{Value: result}, nil
}

func (gs *gatewayServer) getContract(txn *pb.Transaction) (*gateway.Contract, error) {
	wallet := gateway.NewInMemoryWallet()
	wallet.Put("id", gateway.NewX509Identity(txn.Id.Msp, txn.Id.Cert, txn.Id.Key))

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(gs.ccpPath))),
		gateway.WithIdentity(wallet, "id"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to gateway: ")
	}
	defer gw.Close()

	network, err := gw.GetNetwork(txn.Channel)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get network: ")
	}

	return network.GetContract(txn.ChaincodeID), nil
}

func newGatewayServer(ccpPath string) (*gatewayServer, error) {
	return &gatewayServer{ccpPath}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 1234))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	ccpPath := filepath.Join(
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gwServer, _ := newGatewayServer(ccpPath)

	grpcServer := grpc.NewServer()
	pb.RegisterGatewayServer(grpcServer, gwServer)
	//... // determine whether to use TLS
	grpcServer.Serve(lis)
}
