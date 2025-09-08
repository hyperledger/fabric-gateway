#!/usr/bin/env bash

set -eu -o pipefail

TAG="${1:?}"

if [[ ${TAG} != latest ]]; then
    PACKAGE_VERSION=$(jq --raw-output .version package.json)
    NOW=$(date -u '+%Y%m%d.%H%M%S')
    PUBLISH_VERSION="${PACKAGE_VERSION}-dev.${NOW}"
    npm --allow-same-version --no-git-tag-version version "${PUBLISH_VERSION}"
fi

npm publish --access public --tag "${TAG}"
