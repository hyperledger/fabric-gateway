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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-gateway/pkg/connection"
	"github.com/hyperledger/fabric-gateway/pkg/network"
	"github.com/hyperledger/fabric-gateway/pkg/server"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Gateway struct {
		ListenAddress  string   `yaml:"listenAddress"`
		MspId          string   `yaml:"mspId"`
		BootstrapPeers []string `yaml:"bootstrapPeers"`
		Tls            struct {
			Cert struct {
				File string
			}
			Key struct {
				File string
			}
		}
	}
}

func main() {
	host := flag.String("h", "peer0.org1.example.com", "hostname of the bootstrap peer")
	port := flag.Int("p", 7051, "port number of the bootstrap peer")
	mspid := flag.String("m", "", "MSP ID of org")
	tlsPath := flag.String("tlscert", "", "path to the org's TLS Certificate")
	idPath := flag.String("id", "", "path to the gateway's wallet identity")
	certPath := flag.String("cert", "", "path to the gateway's Certificate")
	keyPath := flag.String("key", "", "path to the gateway's private key")

	flag.Parse()

	// read the config file
	cfgFile := os.Getenv("FABRIC_CFG_PATH")
	if cfgFile != "" {
		cfg, err := ioutil.ReadFile(cfgFile + "/gateway.yaml")
		if err != nil {
			log.Printf("No config yaml found at location: %s\n", cfgFile)
		}
		conf := Config{}
		err = yaml.Unmarshal(cfg, &conf)
		if err != nil {
			log.Fatalf("failed to parse gateway config: %s\n", err)
		}
		fmt.Println(conf)
	}

	envCertFile := os.Getenv("GATEWAY_CERT_FILE")
	if envCertFile != "" {
		*certPath = envCertFile
	}

	envKeyFile := os.Getenv("GATEWAY_KEY_FILE")
	if envKeyFile != "" {
		*keyPath = envKeyFile
	}

	envTLSFile := os.Getenv("GATEWAY_TLS_ROOTCERT_FILE")
	if envTLSFile != "" {
		*tlsPath = envTLSFile
	}

	envMspID := os.Getenv("GATEWAY_MSPID")
	if envMspID != "" {
		*mspid = envMspID
	}

	envBootstrapPeers := os.Getenv("GATEWAY_BOOTSTRAP_PEERS")
	if envBootstrapPeers != "" {
		// TODO split up comma separate list
		parts := strings.Split(envBootstrapPeers, ":")
		*host = parts[0]
		*port, _ = strconv.Atoi(parts[1])
	}

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
		bootstrapPeer: &connection.PeerEndpoint{
			Host:    *host,
			Port:    uint32(*port),
			TLSCert: pem,
		},
		mspid: *mspid,
		cert:  cert,
		key:   key,
	}

	// setup server and listen
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 7053))
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	registry, err := network.NewRegistry(config)
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

type bootstrapconfig struct {
	bootstrapPeer *connection.PeerEndpoint
	mspid         string
	cert          string
	key           string
}

func (bc *bootstrapconfig) BootstrapPeer() connection.PeerEndpoint {
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
