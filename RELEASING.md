# Releasing

The following artifacts are created as a result of releasing the Hyperledger Fabric Gateway client API:

- npm package
  - [@hyperledger/fabric-gateway](https://www.npmjs.com/package/@hyperledger/fabric-gateway)
- Java artifact
  - [fabric-gateway](https://central.sonatype.dev/artifact/org.hyperledger.fabric/fabric-gateway/1.0.0/versions)

## Before releasing

The following tasks are required before releasing:

- Update version numbers if required (see below for details).
- Update test, sample, and docs files to match the new version if it was updated.

## Create release

Creating a GitHub release on the [releases page](https://github.com/hyperledger/fabric-gateway/releases) will trigger the build to publish the new release.

When drafting the release, create a new tag for the new version (with a `v` prefix), e.g. `vX.Y.Z`

See previous releases for examples of the title and description.

## After releasing

The following tasks are required after releasing:

- Update version numbers to the next point release (see below for details).
- Update documentation to match the new version, with particular attention to:
  - [docs/compatibility.md](docs/compatibility.md)
  - [java/README.md](java/README.md)

# Versioning

The Hyperledger Fabric Gateway client APIs follow the [Go module version numbering system](https://go.dev/doc/modules/version-numbers)

## Updating version numbers

> **Note:** The [scripts/update-versions.sh](scripts/update-versions.sh) script can be used to perform the updates described below. With no arguments, the script will update to the next patch version. An argument supplied to the script specifies the new version number, which should **not** include a leading `v`.

The following files need to be modified when updating the version number, and these are checked by the build process to ensure they match a tagged release:

- The `GATEWAY_VERSION` variable in [.github/workflows/verify-versions.yml](.github/workflows/verify-versions.yml).
- The `version` element in [java/pom.xml](java/pom.xml).
- The `version` property in [node/package.json](node/package.json).
- The Node package-lock.json files for the Node [implementation](node/package-lock.json) and [scenario tests](scenario/node/package-lock.json).

There is no file to update for the Go SDK, which is versioned by the release tag.

Removing support for Go, Node or Java runtime versions requires at least a minor version change. Adding support for a new runtime version while retaining support for existing versions can be done in a patch release.

## Updating the major version

When updating the major version beyond version 1, Go modules require a new module path. For example, version 2 would require a `/v2` suffix to the module path. This is a disruptive change therefore any incompatible changes which would require a major version change should generally be avoided.

See [Publishing breaking API changes](https://go.dev/doc/modules/release-workflow#breaking) for more details.
