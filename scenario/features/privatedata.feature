#
# Copyright 2021 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Private data collections
    Background:
        Given I have deployed a Fabric network
        And I have created and joined all channels
        And I deploy golang chaincode named private at version 1.0.0 for all organizations on channel mychannel with endorsement policy OR("Org1MSP.member","Org2MSP.member","Org3MSP.member")
        And I create a gateway named gateway1 for user User1 in MSP Org1MSP
        And I connect the gateway to peer0.org1.example.com
        And I use the mychannel network
        And I use the private contract
        And I create a gateway named gateway2 for user User1 in MSP Org2MSP
        And I connect the gateway to peer0.org2.example.com
        And I use the mychannel network
        And I use the private contract
        And I use the gateway named gateway1

    Scenario: Org1 writes private data to Org1Collection and can read it back
        When I prepare to submit a WritePrivateData transaction
        And I set transient data on the transaction to
            | collection | Org1Collection |
            | key | key-101 |
            | value | value-101 |
        And I set the endorsing organizations to ["Org1MSP"]
        And I invoke the transaction
        And I prepare to evaluate a ReadPrivateData transaction
        And I set the transaction arguments to ["Org1Collection", "key-101"]
        And I set the endorsing organizations to ["Org1MSP"]
        And I invoke the transaction
        Then the response should be "value-101"

    Scenario: Org1 writes private data to Org1Collection but fails to read it via an Org2 peer
        When I prepare to submit a WritePrivateData transaction
        And I set transient data on the transaction to
            | collection | Org1Collection |
            | key | key-102 |
            | value | value-102 |
        And I set the endorsing organizations to ["Org1MSP"]
        And I invoke the transaction
        And I prepare to evaluate a ReadPrivateData transaction
        And I set the transaction arguments to ["Org1Collection", "key-102"]
        And I set the endorsing organizations to ["Org2MSP"]
        Then the transaction invocation should fail

    Scenario: Org2 can write private data to Org1Collection (Org1MSP must endorse), but can't read it back
        When I use the gateway named gateway2
        And I prepare to submit a WritePrivateData transaction
        And I set transient data on the transaction to
            | collection | Org1Collection |
            | key | key-103 |
            | value | value-103 |
        And I set the endorsing organizations to ["Org1MSP"]
        And I invoke the transaction
        And I prepare to evaluate a ReadPrivateData transaction
        And I set the transaction arguments to ["Org1Collection", "key-103"]
        And I set the endorsing organizations to ["Org1MSP"]
        Then the transaction invocation should fail

    Scenario: Org2 writes private data to Org1Collection (Org1MSP must endorse), and Org1 can read it
        When I use the gateway named gateway2
        And I prepare to submit a WritePrivateData transaction
        And I set transient data on the transaction to
            | collection | Org1Collection |
            | key | key-104 |
            | value | value-104 |
        And I set the endorsing organizations to ["Org1MSP"]
        And I invoke the transaction
        And I use the gateway named gateway1
        And I prepare to evaluate a ReadPrivateData transaction
        And I set the transaction arguments to ["Org1Collection", "key-104"]
        And I set the endorsing organizations to ["Org1MSP"]
        And I invoke the transaction
        Then the response should be "value-104"

    Scenario: Org1 writes private data to SharedCollection, and Org2 fails to read it
        When I use the gateway named gateway1
        And I prepare to submit a WritePrivateData transaction
        And I set transient data on the transaction to
            | collection | SharedCollection |
            | key | key-106 |
            | value | value-106 |
        And I set the endorsing organizations to ["Org1MSP"]
        And I invoke the transaction
        And I use the gateway named gateway2
        And I prepare to evaluate a ReadPrivateData transaction
        And I set the transaction arguments to ["SharedCollection", "key-106"]
        And I set the endorsing organizations to ["Org1MSP"]
        Then the transaction invocation should fail

    Scenario: Org2 cannot write data to SharedCollection
        When I use the gateway named gateway2
        And I prepare to submit a WritePrivateData transaction
        And I set transient data on the transaction to
            | collection | SharedCollection |
            | key | key-107 |
            | value | value-107 |
        And I set the endorsing organizations to ["Org1MSP"]
        Then the transaction invocation should fail

    # The following scenarios tests the ability for Endorse() to work out which orgs to endorse based on collection policies
    # This one will endorse with the gateway's org
    Scenario: Org1 writes private data to SharedCollection without specifying endorsers
        When I use the gateway named gateway1
        And I prepare to submit a WritePrivateData transaction
        And I set transient data on the transaction to
            | collection | SharedCollection |
            | key | key-108 |
            | value | value-108 |
        And I invoke the transaction
        And I prepare to evaluate a ReadPrivateData transaction
        And I set the transaction arguments to ["SharedCollection", "key-108"]
        And I set the endorsing organizations to ["Org1MSP"]
        And I invoke the transaction
        Then the response should be "value-108"

    # This needs endorsement from Org3
    Scenario: Org1 writes private data to SharedCollection and Org3Collection without specifying endorsers
        When I use the gateway named gateway1
        And I prepare to submit a WritePrivateData transaction
        And I set transient data on the transaction to
            | collection | SharedCollection,Org3Collection |
            | key | key-109 |
            | value | value-109 |
        And I invoke the transaction
        And I prepare to evaluate a ReadPrivateData transaction
        And I set the transaction arguments to ["SharedCollection", "key-109"]
        And I set the endorsing organizations to ["Org1MSP"]
        And I invoke the transaction
        Then the response should be "value-109"
        And I prepare to evaluate a ReadPrivateData transaction
        And I set the transaction arguments to ["Org3Collection", "key-109"]
        And I set the endorsing organizations to ["Org3MSP"]
        And I invoke the transaction
        Then the response should be "value-109"
        And I prepare to evaluate a ReadPrivateData transaction
        And I set the transaction arguments to ["Org3Collection", "key-109"]
        And I set the endorsing organizations to ["Org1MSP"]
        Then the transaction invocation should fail
