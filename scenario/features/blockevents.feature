#
# Copyright 2022 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Block event listening
    Background:
        Given I have deployed a Fabric network
        And I have created and joined all channels
        And I deploy golang chaincode named basic at version 1.0.0 for all organizations on channel mychannel with endorsement policy AND("Org1MSP.member","Org2MSP.member")
        And I create a gateway named mygateway for user User1 in MSP Org1MSP
        And I connect the gateway to peer0.org1.example.com
        And I use the mychannel network
        And I use the basic contract

    Scenario Outline: Receive block event
        When I listen for <type> events
        And I prepare to submit an echo transaction
        And I set the transaction arguments to ["echo"]
        And I invoke the transaction
        Then I should receive a <type> event

        Scenarios:
            | type                   |
            | block                  |
            | filtered block         |
            | block and private data |

    Scenario Outline: Replay block event
        When I prepare to submit an echo transaction
        And I set the transaction arguments to ["echo"]
        And I invoke the transaction
        And I replay <type> events starting at last committed block
        Then I should receive a <type> event

        Scenarios:
            | type                   |
            | block                  |
            | filtered block         |
            | block and private data |

    Scenario Outline: Restart after closing block event session
        Given I listen for <type> events
        And I prepare to submit an echo transaction
        And I set the transaction arguments to ["echo"]
        And I invoke the transaction
        Then I should receive a <type> event
        When I stop listening for <type> events
        And I listen for <type> events
        And I prepare to submit an echo transaction
        And I set the transaction arguments to ["echo"]
        And I invoke the transaction
        Then I should receive a <type> event

        Scenarios:
            | type                   |
            | block                  |
            | filtered block         |
            | block and private data |

    Scenario Outline: Close does not interrupt other block event listeners
        Given I listen for <type> events on a listener named "listener1"
        And I listen for <type> events on a listener named "listener2"
        And I prepare to submit an echo transaction
        And I set the transaction arguments to ["echo"]
        And I invoke the transaction
        Then I should receive a <type> event on "listener1"
        And I should receive a <type> event on "listener2"
        When I stop listening for <type> events on "listener1"
        And I prepare to submit an echo transaction
        And I set the transaction arguments to ["echo"]
        And I invoke the transaction
        Then I should receive a <type> event on "listener2"

        Scenarios:
            | type                   |
            | block                  |
            | filtered block         |
            | block and private data |

    Scenario Outline: Checkpoint of block events
        Given I create a checkpointer
        And I use the checkpointer to listen for <type> events
        And I prepare to submit an echo transaction
        And I set the transaction arguments to ["checkpoint"]
        And I invoke the transaction
        Then I should receive a <type> event
        When I stop listening for <type> events
        And I prepare to submit an echo transaction
        And I set the transaction arguments to ["echo"]
        And I invoke the transaction
        And I use the checkpointer to listen for <type> events
        Then I should receive a <type> event

        Scenarios:
            | type                   |
            | block                  |
            | filtered block         |
            | block and private data |