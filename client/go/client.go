/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/util"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewGatewayClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	idfile := filepath.Join(
		"..",
		"..",
		"..",
		"fabric-samples",
		"fabcar",
		"javascript",
		"wallet",
		"appUser.id",
	)

	id, err := util.ReadWalletIdentity(idfile)
	if err != nil {
		log.Fatalf("failed to read gateway identity: %s", err)
	}

	identity := &pb.Identity{
		Msp:  id.MspID,
		Cert: id.Credentials.Certificate,
		Key:  id.Credentials.Key,
	}

	txn := &pb.Transaction{
		Id:          identity,
		Channel:     "mychannel",
		ChaincodeID: "fabcar",
		TxnName:     "queryAllCars",
		Args:        []string{},
	}

	doit(client.EvaluateTransaction(ctx, txn))

	txn = &pb.Transaction{
		Id:          identity,
		Channel:     "mychannel",
		ChaincodeID: "fabcar",
		TxnName:     "createCar",
		Args:        []string{"CAR10", "VW", "Polo", "Grey", "Mary"},
	}
	doit(client.SubmitTransaction(ctx, txn))

	time.Sleep(5 * time.Second)

	txn.TxnName = "queryCar"
	txn.Args = []string{"CAR10"}
	doit(client.EvaluateTransaction(ctx, txn))

	txn.TxnName = "changeCarOwner"
	txn.Args = []string{"CAR10", "Archie"}
	doit(client.SubmitTransaction(ctx, txn))

	time.Sleep(5 * time.Second)

	txn.TxnName = "queryCar"
	txn.Args = []string{"CAR10"}
	doit(client.EvaluateTransaction(ctx, txn))
}

func doit(result *pb.Result, err error) {
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	fmt.Println(string(result.Value))
}
