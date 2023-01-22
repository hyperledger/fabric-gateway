#!/usr/bin/env sh

# Rename the key files we use to be key.pem instead of a uuid
BASEDIR=$(dirname "$0")

find "${BASEDIR}/crypto-config" -type f -name '*_sk' -print | while read -r KEY; do
    KEY_DIR=$(dirname "${KEY}")
    mv "${KEY}" "${KEY_DIR}/key.pem"
    chmod 644 "${KEY_DIR}/key.pem"
done
