#!/usr/bin/env bash

set -eu

GOFMT=$(gofmt -l "$@")
if [ -n "${GOFMT}" ]; then
    echo 'Go formatting errors:'
    echo "${GOFMT}"
    exit 1
fi
