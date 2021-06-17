#
# Copyright 2021 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Off-line signing
    Background:
        Given I have deployed a Fabric network
        And I have created and joined all channels
        And I deploy golang chaincode named basic at version 1.0.0 for all organizations on channel mychannel with endorsement policy AND("Org1MSP.member","Org2MSP.member")
        And I create a gateway named mygateway without signer for user User1 in MSP Org1MSP
        And I connect the gateway to peer0.org1.example.com
        And I use the mychannel network
        And I use the basic contract

    Scenario: Evaluate fails without signing
        When I prepare to evaluate an echo transaction
        And I set the transaction arguments to ["conga"]
        Then the transaction invocation should fail

    Scenario: Evaluate succeeds with off-line signing
        When I prepare to evaluate an echo transaction
        And I set the transaction arguments to ["conga"]
        And I do off-line signing as user User1 in MSP Org1MSP
        And I invoke the transaction
        Then the response should be "conga"

    Scenario: Submit fails without signing
        When I prepare to submit an echo transaction
        And I set the transaction arguments to ["conga"]
        Then the transaction invocation should fail

    Scenario: Submit succeeds with off-line signing
        When I prepare to submit an echo transaction
        And I set the transaction arguments to ["conga"]
        And I do off-line signing as user User1 in MSP Org1MSP
        And I invoke the transaction
        Then the response should be "conga"
