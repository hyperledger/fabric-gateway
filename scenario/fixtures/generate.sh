#!/usr/bin/env bash

set -eo pipefail

cd docker-compose

docker-compose -f docker-compose-cli.yaml up -d
docker exec cli rm -rf /etc/hyperledger/config/channel.tx \
    /etc/hyperledger/config/core.yaml \
    /etc/hyperledger/config/genesis.block \
    /etc/hyperledger/config/mychannel.block \
    /etc/hyperledger/config/crypto-config
docker exec cli cryptogen generate --config=/etc/hyperledger/config/crypto-config.yaml --output /etc/hyperledger/config/crypto-config
docker exec cli configtxgen -profile ThreeOrgsOrdererGenesis -outputBlock /etc/hyperledger/config/genesis.block -channelID testchainid
docker exec cli configtxgen -profile ThreeOrgsChannel -outputCreateChannelTx /etc/hyperledger/config/channel.tx -channelID mychannel
docker exec cli configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate /etc/hyperledger/config/Org1MSPanchors.tx -channelID mychannel -asOrg Org1MSP
docker exec cli configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate /etc/hyperledger/config/Org2MSPanchors.tx -channelID mychannel -asOrg Org2MSP
docker exec cli configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate /etc/hyperledger/config/Org3MSPanchors.tx -channelID mychannel -asOrg Org3MSP
docker exec cli cp /etc/hyperledger/fabric/core.yaml /etc/hyperledger/config
docker exec cli sh /etc/hyperledger/config/rename_sk.sh
docker-compose -f docker-compose-cli.yaml down --volumes
