#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Discovery
	Background:
		Given I have deployed a tls Fabric network
        And I have created and joined all channels from the tls connection profile
        And I deploy golang chaincode named echo at version 1.0.0 for all organizations on channel mychannel with endorsement policy AND("Org1MSP.member", "Org2MSP.member", "Org3MSP.member")
        And I create a gateway for user User1 in MSP Org1MSP
        And I connect the gateway to localhost:7053
        And I use the mychannel network
        And I use the echo contract

    Scenario: Submit transaction on Go chaincode
        When I stop the peer named peer0.org3.example.com
        And I prepare to submit a addEntry transaction
        And I set the transaction arguments to ["my_name", "my_value"]
        And I invoke the transaction which I expect to fail
        And I start the peer named peer0.org3.example.com
        And I invoke the transaction
        Then the response should equal my_value

#    Scenario: Evaluate transaction on Go chaincode (only peer0.org1 running)
        #When I stop the peer named peer0.org3.example.com
        #And I stop the peer named peer0.org2.example.com
        #And I stop the peer named peer1.org2.example.com
        #And I stop the peer named peer1.org1.example.com
        And I prepare to evaluate a readEntry transaction
        And I set the transaction argument to ["my_name"]
        And I invoke the transaction
        Then the response should equal my_value
