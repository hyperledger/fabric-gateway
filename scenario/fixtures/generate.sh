#!/usr/bin/env bash

set -eu -o pipefail

BASEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"

cd "${BASEDIR}/docker-compose"

docker compose --file docker-compose-cli.yaml run --rm clinopeer /etc/hyperledger/config/create_channel.sh
