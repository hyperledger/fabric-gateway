// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scenario

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

// Default peer counts applied by Fabric to collections that do not specify them.
const (
	defaultRequiredPeerCount = int32(0)
	defaultMaxPeerCount      = int32(1)
)

// collectionConfig describes a private data collection using the JSON representation accepted by the
// "--collections-config" option of the Fabric peer CLI.
type collectionConfig struct {
	Name              string             `json:"name"`
	Policy            string             `json:"policy"`
	RequiredPeerCount *int32             `json:"requiredPeerCount"`
	MaxPeerCount      *int32             `json:"maxPeerCount"`
	BlockToLive       uint64             `json:"blockToLive"`
	MemberOnlyRead    bool               `json:"memberOnlyRead"`
	MemberOnlyWrite   bool               `json:"memberOnlyWrite"`
	EndorsementPolicy *endorsementPolicy `json:"endorsementPolicy"`
}

type endorsementPolicy struct {
	SignaturePolicy     string `json:"signaturePolicy"`
	ChannelConfigPolicy string `json:"channelConfigPolicy"`
}

// readCollectionConfig reads private data collection configuration from a JSON file. No collection configuration is
// returned if the file does not exist.
func readCollectionConfig(file string) (*peer.CollectionConfigPackage, error) {
	content, err := os.ReadFile(file) // #nosec G304
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var configs []collectionConfig
	if err := json.Unmarshal(content, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse collection configuration %s: %w", file, err)
	}

	result := &peer.CollectionConfigPackage{
		Config: make([]*peer.CollectionConfig, 0, len(configs)),
	}

	for _, config := range configs {
		collection, err := config.toProto()
		if err != nil {
			return nil, fmt.Errorf("invalid collection configuration %s: %w", file, err)
		}

		result.Config = append(result.Config, collection)
	}

	return result, nil
}

func (c *collectionConfig) toProto() (*peer.CollectionConfig, error) {
	memberOrgsPolicy, err := newCollectionPolicyConfig(c.Policy)
	if err != nil {
		return nil, err
	}

	endorsement, err := c.EndorsementPolicy.toProto()
	if err != nil {
		return nil, err
	}

	return &peer.CollectionConfig{
		Payload: &peer.CollectionConfig_StaticCollectionConfig{
			StaticCollectionConfig: &peer.StaticCollectionConfig{
				Name:              c.Name,
				MemberOrgsPolicy:  memberOrgsPolicy,
				RequiredPeerCount: valueOrDefault(c.RequiredPeerCount, defaultRequiredPeerCount),
				MaximumPeerCount:  valueOrDefault(c.MaxPeerCount, defaultMaxPeerCount),
				BlockToLive:       c.BlockToLive,
				MemberOnlyRead:    c.MemberOnlyRead,
				MemberOnlyWrite:   c.MemberOnlyWrite,
				EndorsementPolicy: endorsement,
			},
		},
	}, nil
}

func newCollectionPolicyConfig(signaturePolicy string) (*peer.CollectionPolicyConfig, error) {
	applicationPolicy, err := chaincode.NewApplicationPolicy(signaturePolicy, "")
	if err != nil {
		return nil, err
	}

	return &peer.CollectionPolicyConfig{
		Payload: &peer.CollectionPolicyConfig_SignaturePolicy{
			SignaturePolicy: applicationPolicy.GetSignaturePolicy(),
		},
	}, nil
}

func (p *endorsementPolicy) toProto() (*peer.ApplicationPolicy, error) {
	if p == nil {
		return nil, nil
	}

	return chaincode.NewApplicationPolicy(p.SignaturePolicy, p.ChannelConfigPolicy)
}

func valueOrDefault[T any](value *T, defaultValue T) T {
	if value == nil {
		return defaultValue
	}

	return *value
}
