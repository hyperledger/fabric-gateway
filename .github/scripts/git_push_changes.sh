#!/usr/bin/env bash

# Required environment variables
: "${USER_NAME:?}"
: "${USER_EMAIL:?}"
: "${COMMIT_REF:?}"

set -eu -o pipefail

if [ -z "$(git status --porcelain=v1 2>/dev/null)" ]; then
    echo 'No changes to publish'
    exit
fi

git config --local user.name "${USER_NAME}"
git config --local user.email "${USER_EMAIL}"
git add .
git commit -m "Commit ${COMMIT_REF}"
git push
