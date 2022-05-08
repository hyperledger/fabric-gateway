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

    Scenario: Restart after closing chaincode event session
        Given I listen for chaincode events from basic
        And I prepare to submit an event transaction
        And I set the transaction arguments to ["restart", "one"]
        And I invoke the transaction
        Then I should receive a chaincode event named "restart" with payload "one"
        When I stop listening for chaincode events
        And I listen for chaincode events from basic
        And I prepare to submit an event transaction
        And I set the transaction arguments to ["restart", "two"]
        And I invoke the transaction
        Then I should receive a chaincode event named "restart" with payload "two"

    Scenario: Close does not interrupt other chaincode event listeners
        Given I listen for chaincode events from basic on a listener named "listener1"
        And I listen for chaincode events from basic on a listener named "listener2"
        And I prepare to submit an event transaction
        And I set the transaction arguments to ["close", "before"]
        And I invoke the transaction
        Then I should receive a chaincode event named "close" with payload "before" on "listener1"
        And I should receive a chaincode event named "close" with payload "before" on "listener2"
        When I stop listening for chaincode events on "listener1"
        And I prepare to submit an event transaction
        And I set the transaction arguments to ["close", "after"]
        And I invoke the transaction
        Then I should receive a chaincode event named "close" with payload "after" on "listener2"

    Scenario: Checkpoint of chaincode events
        Given I create a checkpointer
        And I use the checkpointer to listen for chaincode events from basic
        And I prepare to submit an event transaction
        And I set the transaction arguments to ["checkpoint", "one"]
        And I invoke the transaction
        Then I should receive a chaincode event named "checkpoint" with payload "one"
        When I stop listening for chaincode events
        And I prepare to submit an event transaction
        And I set the transaction arguments to ["checkpoint", "two"]
        And I invoke the transaction
        And I use the checkpointer to listen for chaincode events from basic
        Then I should receive a chaincode event named "checkpoint" with payload "two"
