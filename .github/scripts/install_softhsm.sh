#!/usr/bin/env bash

# Required environment variables
: "${SOFTHSM2_CONF:?}"
: "${TMPDIR:?}"

set -eu -o pipefail

sudo apt install -y softhsm
echo "directories.tokendir = ${TMPDIR}" > "${SOFTHSM2_CONF}"
softhsm2-util --init-token --slot 0 --label "ForFabric" --pin 98765432 --so-pin 1234
