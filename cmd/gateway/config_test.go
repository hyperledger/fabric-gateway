/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"os"
	"testing"
)

func TestLoadConfigNoOverrides(t *testing.T) {
	os.Setenv("FABRIC_CFG_PATH", "test")
	cfg, err := loadConfig()

	if err != nil {
		t.Fatalf("Failed to load config: %s", err)
	}

	if cfg.MspID() != "Org1MSP" {
		t.Fatalf("Incorrect MSPID value: %s", cfg.MspID())
	}
	if len(cfg.BootstrapPeers()) != 1 && cfg.BootstrapPeers()[0] != "peer0.org1.example.com:7051" {
		t.Fatalf("Incorrect bootstrap peers: %v", cfg.BootstrapPeers())
	}
	if cfg.listenAddress() != "0.0.0.0:7053" {
		t.Fatalf("Incorrect MSPID value: %s", cfg.MspID())
	}
	if cfg.Certificate() == nil {
		t.Fatal("Failed to load certificate")
	}
	if cfg.Key() == nil {
		t.Fatal("Failed to load private key")
	}
	if cfg.TLSRootCert() == nil {
		t.Fatal("Failed to load TLS root certificate")
	}
}

func TestLoadConfigWithOverrides(t *testing.T) {
	os.Setenv("FABRIC_CFG_PATH", "test")
	os.Setenv("GATEWAY_MSPID", "Org2MSP")
	os.Setenv("GATEWAY_BOOTSTRAPPEERS", "peer0.org2.example.com:8051,peer1.org2.example.com:10051")
	cfg, err := loadConfig()

	if err != nil {
		t.Fatalf("Failed to load config: %s", err)
	}

	if cfg.MspID() != "Org2MSP" {
		t.Fatalf("Incorrect MSPID value: %s", cfg.MspID())
	}
	if !(len(cfg.BootstrapPeers()) == 2 &&
		cfg.BootstrapPeers()[0] == "peer0.org2.example.com:8051" &&
		cfg.BootstrapPeers()[1] == "peer1.org2.example.com:10051") {
		t.Fatalf("Incorrect bootstrap peers: %v", cfg.BootstrapPeers())
	}
}
