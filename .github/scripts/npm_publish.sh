#!/usr/bin/env bash

set -eu -o pipefail

TAG="${1:?}"

if [[ ${TAG} != latest ]]; then
    PACKAGE_VERSION=$(jq --raw-output .version package.json)
    TODAY=$(date -u '+%Y%m%d')
    TARGET_VERSION="${PACKAGE_VERSION}-dev.${TODAY}."
    INCREMENT=$(npm view --json | jq --raw-output '.versions[]' | awk -F . "/^${TARGET_VERSION}/"'{ lastVersion=$NF } \
        END { sub(/".*/, "", lastVersion); print (lastVersion=="" ? "1" : lastVersion+1) }')
    PUBLISH_VERSION="${TARGET_VERSION}${INCREMENT}"
    npm --allow-same-version --no-git-tag-version version "${PUBLISH_VERSION}"
fi

npm publish --access public --tag "${TAG}"
