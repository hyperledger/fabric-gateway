#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Discovery
	Background:
		Given I have deployed a tls Fabric network
        And I have created and joined all channels from the tls connection profile
        And I deploy golang chaincode named basic at version 1.0.0 for all organizations on channel mychannel with endorsement policy AND("Org1MSP.member", "Org2MSP.member", "Org3MSP.member")
        And I create a gateway for user User1 in MSP Org1MSP
        And I connect the gateway to localhost:7053
        And I use the mychannel network
        And I use the basic contract

    Scenario: Submit fails with insufficient endorsers
        When I stop the peer named peer0.org3.example.com
        And I prepare to submit an echo transaction
        And I set the transaction arguments to ["conga"]
        Then the transaction invocation should fail

    Scenario: Submit succeeds with sufficient endorsers
        When I prepare to submit an echo transaction
        And I set the transaction arguments to ["conga"]
        And I invoke the transaction
        Then the response should be "conga"

