#!/usr/bin/env bash

set -eu

GOFMT=$(gofmt -l "$@")
if [ ! -z "${GOFMT}" ]; then
    echo 'Go formatting errors:'
    echo "${GOFMT}"
    exit 1
fi
