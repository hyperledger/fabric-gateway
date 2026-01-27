#!/usr/bin/env bash

set -eu -o pipefail

PROJECT_DIR="$( cd "$( dirname -- "${BASH_SOURCE[0]}" )/.." > /dev/null && pwd )"

VERIFY_VERSIONS="${PROJECT_DIR}/.github/workflows/verify-versions.yml"
JAVA_DIR="${PROJECT_DIR}/java"
NODE_DIR="${PROJECT_DIR}/node"

if [ $# -eq 1 ]; then
    NEXT_VERSION="$1"
else
    NEXT_VERSION="$(
        awk -F'GATEWAY_VERSION:' '/^[ \t]*GATEWAY_VERSION:/ {
                sub(/^[ \t]+/, "", $2)
                sub(/[ \t]$/, "", $2)
                split($2, v, ".")
                printf "%d.%d.%d", v[1], v[2], v[3]+1
                exit
            }' \
            "${VERIFY_VERSIONS}"
        )"
fi
NEXT_JAVA_VERSION="${NEXT_VERSION}-SNAPSHOT"

echo "Updating Java version to ${NEXT_JAVA_VERSION}"
( cd "${JAVA_DIR}" && \
    mvn --batch-mode --quiet versions:set -DnewVersion="${NEXT_JAVA_VERSION}" -DgenerateBackupPoms=false )


echo "Updating Node version to ${NEXT_VERSION}"
( cd "${NODE_DIR}" && \
    npm --allow-same-version --no-git-tag-version version "${NEXT_VERSION}" )
make build-scenario-node

echo "Updating verify-versions.yml to ${NEXT_VERSION}"
NEXT_VERIFY_VERSIONS="$( sed -r "s/^([ \t]*GATEWAY_VERSION:[ \t]*)[0-9.]+/\1${NEXT_VERSION}/" "${VERIFY_VERSIONS}" )"
echo "${NEXT_VERIFY_VERSIONS}" > "${VERIFY_VERSIONS}"
