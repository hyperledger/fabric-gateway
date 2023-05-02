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
docker exec cli configtxgen -profile ThreeOrgsApplicationGenesis -outputBlock /etc/hyperledger/config/mychannel.block -channelID mychannel
docker exec cli cp /var/hyperledger/fabric/config/core.yaml /etc/hyperledger/config
docker exec cli /etc/hyperledger/config/rename_sk.sh
docker-compose -f docker-compose-cli.yaml down --volumes
