#!/usr/bin/env bash

set -eu -o pipefail

POM_VERSION=$(mvn org.apache.maven.plugins:maven-help-plugin:evaluate -Dexpression=project.version -q -DforceStdout)
PUBLISH_VERSION="${POM_VERSION%%-*}"

mvn --batch-mode versions:set -DnewVersion="${PUBLISH_VERSION}"
mvn --batch-mode -DskipTests deploy
