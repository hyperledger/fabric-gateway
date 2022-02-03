# Releasing

The following artifacts are created as a result of releasing the Hyperledger Fabric Gateway SDKs:

- npm modules
    - [@hyperledger/fabric-gateway](https://www.npmjs.com/package/@hyperledger/fabric-gateway)
- Java libraries
    - [fabric-gateway](https://search.maven.org/artifact/org.hyperledger.fabric/fabric-gateway)

## Before releasing

The following tasks are required before releasing:

- Update version numbers if required (see below for details)
- Update test, sample, and docs files to match the new version if it was updated

## Create release

Creating a GitHub release on the [releases page](https://github.com/hyperledger/fabric-gateway/releases) will trigger the build to publish the new release.

When drafting the release, create a new tag for the new version (with a `v` prefix), e.g. `vX.Y.Z`

See previous releases for examples of the title and description.

## Publish Java SDK to the Central repository

The automated build process currently only [publishes the Java SDK to Hyperledger's repository](https://hyperledger-fabric.jfrog.io/ui/repos/tree/General/fabric-maven%2Forg%2Fhyperledger%2Ffabric%2Ffabric-gateway).

To publish it to the Central Repository requires an additional manual step (tbc):

```
mvn deploy:deploy-file -DpomFile=<path-to-pom> \
  -Dfile=<path-to-file> \
  -DrepositoryId=<id-to-map-on-server-section-of-settings.xml> \
  -Durl=<url-of-the-repository-to-deploy>
```

## After releasing

The following tasks are required after releasing:

- Update version numbers to the next point release (see below for details)
- Update test, sample, and docs files to match the new version

# Versioning

The Hyperledger Fabric Gateway client APIs follow the [Go module version numbering system](https://go.dev/doc/modules/version-numbers)

## Updating version numbers

The following files need to be modified when updating the version number, and these are checked by the build process to ensure they match a tagged release:

- The `GATEWAY_VERSION` variable in `ci/azure-pipelines.yml`
- The `version` element in `java/pom.xml`
- The `version` property in `node/package.json`

**Note:** there is no file to update for the Go SDK, which is versioned by the release tag.

## Updating the major version

When updating the major version beyond version 1, Go modules require a new module path, e.g. version 2 would require a new `v2` directory containing the new Go module code. This is a very disruptive change therefore any incompatible changes which would require a major version change should be avoided if at all possible.

See [Publishing breaking API changes](https://go.dev/doc/modules/release-workflow#breaking) for more details.
