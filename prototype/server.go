/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/hyperledger/fabric-gateway/pkg/gateway"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()

	// setup server and listen
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 1234))
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	gwServer, _ := gateway.NewGatewayServer()

	grpcServer := grpc.NewServer()
	pb.RegisterGatewayServer(grpcServer, gwServer)
	//... // determine whether to use TLS
	grpcServer.Serve(lis)
}
