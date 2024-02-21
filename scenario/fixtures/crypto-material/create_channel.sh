#!/usr/bin/env bash

set -eu -o pipefail

BASEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"

clean_config() {
    rm -rf /etc/hyperledger/config/channel.tx \
        /etc/hyperledger/config/core.yaml \
        /etc/hyperledger/config/genesis.block \
        /etc/hyperledger/config/mychannel.block \
        /etc/hyperledger/config/crypto-config
}

create_channel_config() {
    cryptogen generate --config=/etc/hyperledger/config/crypto-config.yaml \
        --output=/etc/hyperledger/config/crypto-config
    configtxgen -configPath /etc/hyperledger/config \
        -profile ThreeOrgsApplicationGenesis \
        -outputBlock /etc/hyperledger/config/mychannel.block \
        -channelID mychannel
}

rename_secret_key() {
    local KEY
    local KEY_DIR

    find "${BASEDIR}/crypto-config" -type f -name '*_sk' -print | while read -r KEY; do
        KEY_DIR=$(dirname "${KEY}")
        mv "${KEY}" "${KEY_DIR}/key.pem"
        chmod 644 "${KEY_DIR}/key.pem"
    done
}

clean_config
create_channel_config
rename_secret_key

cp "${FABRIC_CFG_PATH}/core.yaml" /etc/hyperledger/config
