#
# Copyright 2021 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Chaincode event listening
    Background:
        Given I have deployed a Fabric network
        And I have created and joined all channels
        And I deploy golang chaincode named basic at version 1.0.0 for all organizations on channel mychannel with endorsement policy AND("Org1MSP.member","Org2MSP.member")
        And I create a gateway named mygateway for user User1 in MSP Org1MSP
        And I connect the gateway to peer0.org1.example.com
        And I use the mychannel network
        And I use the basic contract

    Scenario: Receive chaincode event
        When I listen for chaincode events from basic
        And I prepare to submit an event transaction
        And I set the transaction arguments to ["foo", "bar"]
        And I invoke the transaction
        Then I should receive a chaincode event named "foo" with payload "bar"
    
    Scenario: Replay chaincode event
        When I prepare to submit an event transaction
        And I set the transaction arguments to ["event", "replay"]
        And I invoke the transaction
        And I replay chaincode events from basic starting at last committed block
        Then I should receive a chaincode event named "event" with payload "replay"
