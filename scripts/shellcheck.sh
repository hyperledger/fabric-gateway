#!/usr/bin/env bash

fileCount=0
failCount=0

while read -r scriptFile; do
    if [ -n "${scriptFile}" ]; then
    ((fileCount++))
        shellcheck "${scriptFile}" || ((failCount++))
    fi
done <<< "$(find . -type d \( -name vendor -o -name node_modules \) -prune -o -type f -name '*.sh' -print)"

echo "Checked ${fileCount} files with ${failCount} failing."
exit "${failCount}"
