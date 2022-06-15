/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scenario

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages-go/v16"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type TransactionType int

const (
	Evaluate TransactionType = iota
	Submit
)

const defaultListenerName = ""

var (
	gateways                 map[string]*GatewayConnection
	currentGateway           *GatewayConnection
	transaction              *Transaction
	lastCommittedBlockNumber uint64
)

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.AfterSuite(func() {
		if os.Getenv("GATEWAY_NO_SHUTDOWN") != "TRUE" {
			stopFabric()
		}
	})
}

func InitializeScenario(s *godog.ScenarioContext) {
	s.Before(beforeScenario)
	s.After(afterScenario)

	s.Step(`^I register and enroll an HSM user (\S+) in MSP Org1MSP$`, generateHSMUser)
	s.Step(`^I create a gateway named (\S+) for user (\S+) in MSP (\S+)$`, createGateway)
	s.Step(`^I create a gateway named (\S+) for HSM user (\S+) in MSP (\S+)`, createGatewayWithHSMSigner)
	s.Step(`^I create a gateway named (\S+) without signer for user (\S+) in MSP (\S+)$`, createGatewayWithoutSigner)
	s.Step(`^I create a checkpointer`, createCheckpointer)
	s.Step(`^I connect the gateway to (\S+)$`, connectGateway)
	s.Step(`^I use the gateway named (\S+)$`, useGateway)
	s.Step(`^I create a checkpointer`, createCheckpointer)
	s.Step(`^I deploy (\S+) chaincode named (\S+) at version (\S+) for all organizations on channel (\S+) with endorsement policy (.+)$`, deployChaincode)
	s.Step(`^I have created and joined all channels$`, createAndJoinChannels)
	s.Step(`^I have deployed a Fabric network$`, haveFabricNetwork)
	s.Step(`^I prepare to (submit|evaluate) an? (\S+) transaction$`, prepareTransaction)
	s.Step(`^I set the transaction arguments? to (.+)$`, setArguments)
	s.Step(`^I set transient data on the transaction to$`, setTransientData)
	s.Step(`^I set the endorsing organizations? to (.+)$`, setEndorsingOrgs)
	s.Step(`^I do off-line signing as user (\S+) in MSP (\S+)$`, useOfflineSigner)
	s.Step(`^I invoke the transaction$`, invokeSuccessfulTransaction)
	s.Step(`^I use the (\S+) contract$`, useContract)
	s.Step(`^I use the (\S+) network$`, useNetwork)
	s.Step(`^I use the checkpointer to listen for chaincode events from (\S+)$`, listenForChaincodeEventsUsingCheckpointer)
	s.Step(`^the response should be JSON matching$`, theResponseShouldBeJSONMatching)
	s.Step(`^I stop the peer named (\S+)$`, stopPeer)
	s.Step(`^I start the peer named (\S+)$`, startPeer)
	s.Step(`^the response should be "([^"]*)"$`, theResponseShouldBe)
	s.Step(`^the transaction invocation should fail$`, theTransactionShouldFail)
	s.Step(`^the error message should contain "([^"]*)"$`, theErrorMessageShouldContain)
	s.Step(`^the error details should be$`, theErrorDetailsShouldBe)
	s.Step(`^the error status should be (\S+)$`, theErrorStatusShouldBe)
	s.Step(`^I listen for chaincode events from (\S+)$`, listenForChaincodeEvents)
	s.Step(`^I listen for chaincode events from (\S+) on a listener named "([^"]*)"$`, listenForChaincodeEventsOnListener)
	s.Step(`^I replay chaincode events from (\S+) starting at last committed block$`, replayChaincodeEventsFromLastBlock)
	s.Step(`^I stop listening for chaincode events$`, stopChaincodeEventListening)
	s.Step(`^I stop listening for chaincode events on "([^"]*)"$`, stopChaincodeEventListeningOnListener)
	s.Step(`^I use the checkpointer to listen for block events$`, listenForBlockEventsUsingCheckpointer)
	s.Step(`^I use the checkpointer to listen for filtered block events$`, listenForFilteredBlockEventsUsingCheckpointer)
	s.Step(`^I use the checkpointer to listen for block and private data events$`, listenForBlockAndPrivateDataUsingCheckpointer)
	s.Step(`^I should receive a chaincode event named "([^"]*)" with payload "([^"]*)"$`, receiveChaincodeEvent)
	s.Step(`^I should receive a chaincode event named "([^"]*)" with payload "([^"]*)" on "([^"]*)"$`, receiveChaincodeEventOnListener)
	s.Step(`^I listen for block events$`, listenForBlockEvents)
	s.Step(`^I listen for block events on a listener named "([^"]*)"$`, listenForBlockEventsOnListener)
	s.Step(`^I replay block events starting at last committed block$`, replayBlockEventsFromLastBlock)
	s.Step(`^I stop listening for block events$`, stopBlockEventListening)
	s.Step(`^I stop listening for block events on "([^"]*)"$`, stopBlockEventListeningOnListener)
	s.Step(`^I should receive a block event$`, receiveBlockEvent)
	s.Step(`^I should receive a block event on "([^"]*)"$`, receiveBlockEventOnListener)
	s.Step(`^I listen for filtered block events$`, listenForFilteredBlockEvents)
	s.Step(`^I listen for filtered block events on a listener named "([^"]*)"$`, listenForFilteredBlockEventsOnListener)
	s.Step(`^I replay filtered block events starting at last committed block$`, replayFilteredBlockEventsFromLastBlock)
	s.Step(`^I stop listening for filtered block events$`, stopFilteredBlockEventListening)
	s.Step(`^I stop listening for filtered block events on "([^"]*)"$`, stopFilteredBlockEventListeningOnListener)
	s.Step(`^I should receive a filtered block event$`, receiveFilteredBlockEvent)
	s.Step(`^I should receive a filtered block event on "([^"]*)"$`, receiveFilteredBlockEventOnListener)
	s.Step(`^I listen for block and private data events$`, listenForBlockAndPrivateDataEvents)
	s.Step(`^I listen for block and private data events on a listener named "([^"]*)"$`, listenForBlockAndPrivateDataEventsOnListener)
	s.Step(`^I replay block and private data events starting at last committed block$`, replayBlockAndPrivateDataEventsFromLastBlock)
	s.Step(`^I stop listening for block and private data events$`, stopBlockAndPrivateDataEventListening)
	s.Step(`^I stop listening for block and private data events on "([^"]*)"$`, stopBlockAndPrivateDataEventListeningOnListener)
	s.Step(`^I should receive a block and private data event$`, receiveBlockAndPrivateDataEvent)
	s.Step(`^I should receive a block and private data event on "([^"]*)"$`, receiveBlockAndPrivateDataEventOnListener)
}

func beforeScenario(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	gateways = make(map[string]*GatewayConnection)
	currentGateway = nil
	transaction = nil
	return ctx, nil
}

func afterScenario(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	for _, connection := range gateways {
		connection.Close()
	}
	return ctx, nil
}

func createGateway(name string, user string, mspID string) error {
	connection, err := NewGatewayConnectionWithSigner(user, mspID)
	if err != nil {
		return err
	}

	currentGateway = connection
	gateways[name] = connection
	return nil
}

func createGatewayWithHSMSigner(name string, user string, mspID string) error {
	connection, err := NewGatewayConnectionWithHSMSigner(user, mspID)
	if err != nil {
		return err
	}

	currentGateway = connection
	gateways[name] = connection
	return nil
}

func createGatewayWithoutSigner(name string, user string, mspID string) error {
	connection, err := NewGatewayConnection(user, mspID, false)
	if err != nil {
		return err
	}

	currentGateway = connection
	gateways[name] = connection
	return nil
}

func createCheckpointer() {
	currentGateway.createCheckpointer()
}

func connectGateway(peer string) error {
	conn, ok := peerConnectionInfos[peer]
	if !ok {
		return fmt.Errorf("no connection info found for peer: %s", peer)
	}

	certificate, err := loadX509Cert(conn.tlsRootCertPath)
	if err != nil {
		return err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)

	url := conn.host + ":" + strconv.FormatUint(uint64(conn.port), 10)

	transportCredentials := credentials.NewClientTLSFromCert(certPool, conn.serverNameOverride)
	clientConn, err := grpc.Dial(url, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return err
	}

	return currentGateway.Connect(clientConn)
}

func useGateway(name string) error {
	var ok bool
	if currentGateway, ok = gateways[name]; !ok {
		return fmt.Errorf("no gateway found: %s", name)
	}
	return nil
}

func loadX509Cert(certFile string) (*x509.Certificate, error) {
	cf, e := ioutil.ReadFile(certFile)
	if e != nil {
		return nil, e
	}

	cpb, _ := pem.Decode(cf)
	crt, e := x509.ParseCertificate(cpb.Bytes)

	if e != nil {
		return nil, e
	}
	return crt, nil
}

func prepareTransaction(action string, name string) error {
	txType := Submit
	if action == "evaluate" {
		txType = Evaluate
	}

	tx, err := currentGateway.PrepareTransaction(txType, name)
	transaction = tx
	return err
}

func setArguments(argsJSON string) error {
	args, err := unmarshalArgs(argsJSON)
	if err != nil {
		return err
	}

	transaction.AddOptions(client.WithArguments(args...))
	return nil
}

func useOfflineSigner(user string, mspID string) error {
	keyPath := PrivateKeyPath(user, mspID)
	sign, err := NewSign(keyPath)
	if err != nil {
		return err
	}

	transaction.SetOfflineSign(sign)
	return nil
}

func unmarshalArgs(argsJSON string) ([]string, error) {
	var args []string
	err := json.Unmarshal([]byte(argsJSON), &args)
	if err != nil {
		return nil, err
	}

	return args, nil
}

func setTransientData(table *messages.PickleTable) error {
	transient := make(map[string][]byte)
	for _, row := range table.Rows {
		transient[row.Cells[0].Value] = []byte(row.Cells[1].Value)
	}

	transaction.AddOptions(client.WithTransient(transient))
	return nil
}

func setEndorsingOrgs(argsJSON string) error {
	args, err := unmarshalArgs(argsJSON)
	if err != nil {
		return err
	}

	transaction.AddOptions(client.WithEndorsingOrganizations(args...))
	return nil
}

func invokeSuccessfulTransaction() error {
	if err := invokeTransaction(); err != nil {
		if s, ok := status.FromError(err); ok {
			fmt.Printf("Error details: %+v\n", s.Details())
		}
		return err
	}

	return nil
}

func invokeTransaction() error {
	err := transaction.Invoke()
	lastCommittedBlockNumber = transaction.BlockNumber()
	return err
}

func useNetwork(channelName string) error {
	return currentGateway.UseNetwork(channelName)
}

func useContract(contractName string) error {
	return currentGateway.UseContract(contractName)
}

func theTransactionShouldFail() error {
	err := invokeTransaction()
	if err == nil {
		return fmt.Errorf("transaction invocation was expected to fail, but it returned: %s", transaction.Result())
	}
	if s, ok := status.FromError(err); ok {
		fmt.Printf("Error details: %+v\n", s.Details())
	}
	return nil
}

func theResponseShouldBeJSONMatching(arg *messages.PickleDocString) error {
	same, err := jsonEqual([]byte(arg.Content), transaction.Result())
	if err != nil {
		return err
	}
	if !same {
		return fmt.Errorf("transaction response doesn't match expected value. Got: %s", transaction.Result())
	}
	return nil
}

func jsonEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}

func theResponseShouldBe(expected string) error {
	actual := string(transaction.Result())
	if actual != expected {
		return fmt.Errorf("transaction response \"%s\" does not match expected value \"%s\"", actual, expected)
	}
	return nil
}

func theErrorMessageShouldContain(expected string) error {
	transactionErr := transaction.Err()
	if transactionErr == nil {
		return fmt.Errorf("no transaction error, result is %s", string(transaction.Result()))
	}

	actual := transactionErr.Error()
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("transaction error message \"%s\" does not contain expected value \"%s\"", actual, expected)
	}

	return nil
}

func theErrorDetailsShouldBe(table *messages.PickleTable) error {
	details := transaction.ErrDetails()
	expected := map[string]*gateway.ErrorDetail{}
	for _, row := range table.Rows {
		address := row.Cells[0].Value
		mspid := row.Cells[1].Value
		msg := row.Cells[2].Value
		expected[address] = &gateway.ErrorDetail{
			MspId:   mspid,
			Address: address,
			Message: msg,
		}
	}
	for _, detail := range details {
		ee := expected[detail.Address]
		if ee == nil {
			return fmt.Errorf("unexpected error from endpoint: %s", detail.Address)
		}
		if !strings.Contains(detail.Message, ee.Message) {
			return fmt.Errorf("expected error detail %+v, got %+v", ee, detail)
		}
		delete(expected, detail.Address)
	}
	if len(expected) > 0 {
		keys := make([]string, 0, len(expected))
		for k := range expected {
			keys = append(keys, k)
		}
		return fmt.Errorf("expected error details from endpoint(s): %v", keys)
	}
	return nil
}

func theErrorStatusShouldBe(expected string) error {
	var expectedCode codes.Code
	if err := expectedCode.UnmarshalJSON([]byte("\"" + expected + "\"")); err != nil {
		return err
	}

	actual := status.Code(transaction.Err())
	if actual != expectedCode {
		return fmt.Errorf("expected status %v, got %v: %w", expectedCode, actual, transaction.Err())
	}

	return nil
}

func listenForChaincodeEvents(chaincodeName string) error {
	return listenForChaincodeEventsOnListener(chaincodeName, defaultListenerName)
}

func listenForChaincodeEventsUsingCheckpointer(chaincodeName string) error {
	return currentGateway.ListenForChaincodeEventsUsingCheckpointer(defaultListenerName, chaincodeName)
}

func listenForChaincodeEventsOnListener(chaincodeName string, listenerName string) error {
	return currentGateway.ListenForChaincodeEvents(listenerName, chaincodeName)
}

func replayChaincodeEventsFromLastBlock(chaincodeName string) error {
	return currentGateway.ReplayChaincodeEvents(defaultListenerName, chaincodeName, lastCommittedBlockNumber)
}

func stopChaincodeEventListening() {
	stopChaincodeEventListeningOnListener(defaultListenerName)
}

func stopChaincodeEventListeningOnListener(listenerName string) {
	currentGateway.CloseChaincodeEvents(listenerName)
}

func receiveChaincodeEvent(name string, payload string) error {
	return receiveChaincodeEventOnListener(name, payload, defaultListenerName)
}

func receiveChaincodeEventOnListener(name string, payload string, listenerName string) error {
	event, err := currentGateway.ChaincodeEvent(listenerName)
	if err != nil {
		return err
	}

	if event.EventName != name || string(event.Payload) != payload {
		return fmt.Errorf("expected event named \"%s\" with payload \"%s\", got: %v", name, payload, event)
	}
	return nil
}

func listenForBlockEvents() error {
	return listenForBlockEventsOnListener(defaultListenerName)
}

func listenForBlockEventsOnListener(name string) error {
	return currentGateway.ListenForBlockEvents(name)
}

func listenForBlockEventsUsingCheckpointer() error {
	return currentGateway.ListenForBlockEventsUsingCheckpointer(defaultListenerName)
}

func listenForFilteredBlockEventsUsingCheckpointer() error {
	return currentGateway.ListenForFilteredBlockEventsUsingCheckpointer(defaultListenerName)
}

func listenForBlockAndPrivateDataUsingCheckpointer() error {
	return currentGateway.ListenForBlockAndPrivateDataEventsUsingCheckpointer(defaultListenerName)
}

func replayBlockEventsFromLastBlock() error {
	return currentGateway.ReplayBlockEvents(defaultListenerName, lastCommittedBlockNumber)
}

func stopBlockEventListening() {
	stopBlockEventListeningOnListener(defaultListenerName)
}

func stopBlockEventListeningOnListener(listenerName string) {
	currentGateway.CloseBlockEvents(listenerName)
}

func receiveBlockEvent() error {
	return receiveBlockEventOnListener(defaultListenerName)
}

func receiveBlockEventOnListener(listenerName string) error {
	_, err := currentGateway.BlockEvent(listenerName)

	if err != nil {
		return err
	}

	return nil
}

func listenForFilteredBlockEvents() error {
	return listenForFilteredBlockEventsOnListener(defaultListenerName)
}

func listenForFilteredBlockEventsOnListener(name string) error {
	return currentGateway.ListenForFilteredBlockEvents(name)
}

func replayFilteredBlockEventsFromLastBlock() error {
	return currentGateway.ReplayFilteredBlockEvents(defaultListenerName, lastCommittedBlockNumber)
}

func stopFilteredBlockEventListening() {
	stopBlockEventListeningOnListener(defaultListenerName)
}

func stopFilteredBlockEventListeningOnListener(listenerName string) {
	currentGateway.CloseFilteredBlockEvents(listenerName)
}

func receiveFilteredBlockEvent() error {
	return receiveFilteredBlockEventOnListener(defaultListenerName)
}

func receiveFilteredBlockEventOnListener(listenerName string) error {
	_, err := currentGateway.FilteredBlockEvent(listenerName)
	if err != nil {
		return err
	}

	return nil
}

func listenForBlockAndPrivateDataEvents() error {
	return listenForBlockAndPrivateDataEventsOnListener(defaultListenerName)
}

func listenForBlockAndPrivateDataEventsOnListener(name string) error {
	return currentGateway.ListenForBlockAndPrivateDataEvents(name)
}

func replayBlockAndPrivateDataEventsFromLastBlock() error {
	return currentGateway.ReplayBlockAndPrivateDataEvents(defaultListenerName, lastCommittedBlockNumber)
}

func stopBlockAndPrivateDataEventListening() {
	stopBlockAndPrivateDataEventListeningOnListener(defaultListenerName)
}

func stopBlockAndPrivateDataEventListeningOnListener(listenerName string) {
	currentGateway.CloseBlockAndPrivateDataEvents(listenerName)
}

func receiveBlockAndPrivateDataEvent() error {
	return receiveBlockAndPrivateDataEventOnListener(defaultListenerName)
}

func receiveBlockAndPrivateDataEventOnListener(listenerName string) error {
	_, err := currentGateway.BlockAndPrivateDataEvent(listenerName)
	if err != nil {
		return err
	}
	return nil
}
