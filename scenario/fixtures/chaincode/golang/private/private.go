/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Fatalf("Error creating chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Fatalf("Error starting chaincode: %v", err)
	}
}

// SmartContract provides the contract implementation
type SmartContract struct {
	contractapi.Contract
}

// GetPeerOrg returns the mspid of the current peer
func (s *SmartContract) GetPeerOrg(_ contractapi.TransactionContextInterface) (string, error) {
	peerOrgID, err := shim.GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting client's orgID: %v", err)
	}

	return peerOrgID, nil
}

// CheckEndorsingOrg checks that the peer org is present in the given transient data
func (s *SmartContract) CheckEndorsingOrg(ctx contractapi.TransactionContextInterface) (string, error) {
	transient, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("failed to get transient data: %w", err)
	}

	peerOrgID, err := shim.GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting client's orgID: %v", err)
	}
	// collection := fmt.Sprintf("_implicit_org_%s", peerOrgID)

	if _, ok := transient[peerOrgID]; ok {
		return "success", nil
	}

	return "", fmt.Errorf("endorser in this org (%s) should not have been invoked", peerOrgID)
}

// WritePrivateData writes the transient data to private data collection(s)
func (s *SmartContract) WritePrivateData(ctx contractapi.TransactionContextInterface) (string, error) {
	transient, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("failed to get transient data: %w", err)
	}

	c, ok := transient["collection"]
	if !ok {
		return "", fmt.Errorf("transient data doesn't contain collection name")
	}
	collections := strings.Split(string(c), ",")
	key, ok := transient["key"]
	if !ok {
		return "", fmt.Errorf("transient data doesn't contain key")
	}
	value, ok := transient["value"]
	if !ok {
		return "", fmt.Errorf("transient data doesn't contain value")
	}

	for _, collection := range collections {
		err = ctx.GetStub().PutPrivateData(collection, string(key), value)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

// ReadPrivateData attempts to read from a private data collection
func (s *SmartContract) ReadPrivateData(ctx contractapi.TransactionContextInterface, collection, key string) (string, error) {
	value, err := ctx.GetStub().GetPrivateData(collection, key)
	return string(value), err
}

func (s *SmartContract) SetStateWithEndorser(ctx contractapi.TransactionContextInterface, key, value, endorser string) (string, error) {
	err := ctx.GetStub().PutState(key, []byte(value))
	if err != nil {
		return "", err
	}
	return "", setStateEndorser(ctx, key, endorser)
}

func (s *SmartContract) SetStateEndorsers(ctx contractapi.TransactionContextInterface, key, endorser1, endorser2 string) (string, error) {
	return "", setStateEndorser(ctx, key, endorser1, endorser2)
}

func (s *SmartContract) GetState(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	value, err := ctx.GetStub().GetState(key)
	return string(value), err
}

func (s *SmartContract) ChangeState(ctx contractapi.TransactionContextInterface, key, value string) (string, error) {
	return "", ctx.GetStub().PutState(key, []byte(value))
}

func setStateEndorser(ctx contractapi.TransactionContextInterface, key string, endorsers ...string) error {
	endorsementPolicy, err := statebased.NewStateEP(nil)
	if err != nil {
		return err
	}
	for _, endorser := range endorsers {
		err = endorsementPolicy.AddOrgs(statebased.RoleTypePeer, endorser)
		if err != nil {
			return fmt.Errorf("failed to add org to endorsement policy: %v", err)
		}
	}
	policy, err := endorsementPolicy.Policy()
	if err != nil {
		return fmt.Errorf("failed to create endorsement policy bytes from org: %v", err)
	}
	return ctx.GetStub().SetStateValidationParameter(key, policy)
}
