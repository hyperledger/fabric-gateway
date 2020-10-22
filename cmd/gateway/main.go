/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"log"
	"net"

	"github.com/hyperledger/fabric-gateway/pkg/network"
	"github.com/hyperledger/fabric-gateway/pkg/server"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"google.golang.org/grpc"
)

func main() {
	// read the config file
	conf, err := loadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	// setup server and listen
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 7053))
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	registry, err := network.NewRegistry(conf)
	if err != nil {
		log.Fatalf("failed to create network registry: %s", err)
	}
	gwServer, _ := server.NewGatewayServer(registry)

	grpcServer := grpc.NewServer()
	pb.RegisterGatewayServer(grpcServer, gwServer)
	//... // determine whether to use TLS
	fmt.Println("Gateway listening and serving")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
