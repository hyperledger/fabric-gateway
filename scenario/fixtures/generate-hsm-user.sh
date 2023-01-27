#!/usr/bin/env bash

set -eo pipefail

# Required environment variables
: "${SOFTHSM2_CONF:?}"

# define the CA setup
CA_URL="127.0.0.1:7054"

# try to locate the Soft HSM library
POSSIBLE_LIB_LOC=(
    '/usr/lib/softhsm/libsofthsm2.so'
    '/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so'
    '/usr/local/lib/softhsm/libsofthsm2.so'
    '/usr/lib/libacsp-pkcs11.so'
    '/opt/homebrew/lib/softhsm/libsofthsm2.so'
)
for TEST_LIB in "${POSSIBLE_LIB_LOC[@]}"; do
    if [ -f "${TEST_LIB}" ]; then
        HSM2_LIB="${TEST_LIB}"
        break
    fi
done
[ -z "${HSM2_LIB}" ] && echo No SoftHSM PKCS11 Library found, ensure you have installed softhsm2 && exit 1

# Update the client config file to point to the softhsm pkcs11 library
CLIENT_CONFIG=./ca-client-config/fabric-ca-client-config.yaml
sed "s+REPLACE_ME_HSMLIB+${HSM2_LIB}+g" < ca-client-config/fabric-ca-client-config-template.yaml > "${CLIENT_CONFIG}"

# create the users, remove any existing users
CRYPTO_PATH="${PWD}/crypto-material/hsm"
[ -d "${CRYPTO_PATH}" ] && rm -fr "${CRYPTO_PATH}"

CA_ADMIN='admin'
CA_ADMIN_PW='adminpw'
# user passed in as parameter
HSM_USER="$1"

fabric-ca-client enroll \
    -c "${CLIENT_CONFIG}" \
    -u "http://${CA_ADMIN}:${CA_ADMIN_PW}@${CA_URL}" \
    --mspdir "${CRYPTO_PATH}/${CA_ADMIN}" \
    --csr.hosts example.com

! fabric-ca-client register \
    -c "${CLIENT_CONFIG}" \
    --mspdir "${CRYPTO_PATH}/${CA_ADMIN}" \
    --id.name "${HSM_USER}" \
    --id.secret "${HSM_USER}" \
    --id.type client \
    --caname ca-org1 \
    --id.maxenrollments 0 \
    -m example.com \
    -u "http://${CA_URL}" && echo 'user probably already registered, continuing'

fabric-ca-client enroll \
    -c "${CLIENT_CONFIG}" \
    -u "http://${HSM_USER}:${HSM_USER}@${CA_URL}" \
    --mspdir "${CRYPTO_PATH}/${HSM_USER}" \
    --csr.hosts example.com
