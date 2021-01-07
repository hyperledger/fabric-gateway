/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package network

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
)

func (reg *registry) ListenForTxEvents(channel string, txid string, done chan<- bool) error {
	envelope, err := createDeliverEnvelope(channel, reg.signer)
	if err != nil {
		return fmt.Errorf("Failed to create deliver env: %w", err)
	}
	eventCh := make(chan *peer.FilteredTransaction)

	deliverClients := reg.GetDeliverers(channel)
	for _, client := range deliverClients {
		go listenForTxEvent(client, txid, envelope, eventCh)
	}

	for i := 0; i < len(deliverClients); i++ {
		select {
		case ev := <-eventCh:
			fmt.Println("received", ev)
		}
	}
	done <- true
	return nil
}

func createDeliverEnvelope(
	channelID string,
	// certificate tls.Certificate,
	signer *signingIdentity,
) (*common.Envelope, error) {
	// var tlsCertHash []byte
	// check for client certificate and create hash if present
	// if len(certificate.Certificate) > 0 {
	// 	tlsCertHash = util.ComputeSHA256(certificate.Certificate[0])
	// }

	start := &ab.SeekPosition{
		Type: &ab.SeekPosition_Newest{
			Newest: &ab.SeekNewest{},
		},
	}

	stop := &ab.SeekPosition{
		Type: &ab.SeekPosition_Specified{
			Specified: &ab.SeekSpecified{
				Number: math.MaxUint64,
			},
		},
	}

	seekInfo := &ab.SeekInfo{
		Start:    start,
		Stop:     stop,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}

	env, err := protoutil.CreateSignedEnvelope(
		common.HeaderType_DELIVER_SEEK_INFO,
		channelID,
		signer,
		seekInfo,
		int32(0),
		uint64(0),
		// tlsCertHash,
	)
	if err != nil {
		return nil, fmt.Errorf("Error signing envelope: %w", err)
	}

	return env, nil
}

func listenForTxEvent(deliverClient peer.DeliverClient, txid string, envelope *common.Envelope, events chan<- *peer.FilteredTransaction) error {
	// listen for events
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	df, err := deliverClient.DeliverFiltered(ctx)
	if err != nil {
		return fmt.Errorf("Failed to register for events: %w", err)
	}
	defer df.CloseSend()
	err = df.Send(envelope)
	if err != nil {
		return fmt.Errorf("Error sending deliver seek info envelope: %w", err)
	}

	for {
		resp, err := df.Recv()
		if err != nil {
			return fmt.Errorf("error receiving from deliver filtered: %w", err)
		}
		switch r := resp.Type.(type) {
		case *peer.DeliverResponse_FilteredBlock:
			filteredTransactions := r.FilteredBlock.FilteredTransactions
			for _, tx := range filteredTransactions {
				if tx.Txid == txid {
					events <- tx
					return nil
				}
			}
		case *peer.DeliverResponse_Status:
			return fmt.Errorf("deliver completed with status (%s) before txid received", r.Status)
		default:
			return fmt.Errorf("received unexpected response type (%T)", r)
		}
	}
}
