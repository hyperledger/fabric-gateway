# Contributing to Hyperledger Fabric Gateway

We welcome contributions to this project and to [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io) projects in general. There is always plenty to do!

If you have any questions about the project or how to contribute, you can use:

- The GitHub repository [Discussions](https://github.com/hyperledger/fabric-gateway/discussions).
- The `#fabric-client-apis` channel on Hyperledger [Discord](https://discord.com/channels/905194001349627914/943089887589048350) ([invite link](https://discord.gg/hyperledger)).
- The [Fabric mailing list](https://lists.hyperledger.org/g/fabric).

Here are a few guidelines to help you contribute successfully...

## Issues

Issues are tracked in the GitHub repository [Issues](https://github.com/hyperledger/fabric-gateway/issues). *Please do not use issues to ask questions.*

If you find a bug which we don't already know about, you can help us by creating a new issue describing the problem. Please include as much detail as possible to help us track down the cause.

## Enhancements

If you have a proposal for new functionality, either for the community to consider or that you would like to contribute yourself, please first raise an issue describing this functionality. Make the title something reasonably short but descriptive, and include a [user story](https://en.wikipedia.org/wiki/User_story) description of the enhancement, followed by any supporting information. For example, [issue #198](https://github.com/hyperledger/fabric-gateway/issues/198). The *"So that"* statement provides useful context about the motivation for the enhancement, and helps in determining whether any changes successfully satisfy the requirement.

*Make sure you have the support of the maintainers and community before investing a lot of effort in project enhancements.*

## Contributing code

If you want to begin contributing code, looking through our open issues is a good way to start. Try looking for issues with `help wanted` or `good first issue` tags first, or ask us on Discord if you're unsure which issue to choose.

Code changes should include code, tests and documentation. This includes scenario tests. For larger enhancements, it may be appropriate to deliver language implementations and scenario tests in separate pull requests.

Commits must include a **Signed-off-by** trailer, as produced by `git commit --signoff`.

When creating a pull request, include a reference to the issue that it contributes to or fixes, as described in the [GitHub documentation](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue). Pull requests should generally contain a single commit with a commit message that describes the change, and any supporting explanation of the implementation or reasons for the change. If the pull request review requires changes to be made, these can be delivered as additional commits.

It is helpful to create work-in-progress pull requests as [draft](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests#draft-pull-requests), and only mark them as ready for review when you feel the change is complete.

### Go

Go code uses [testify](https://github.com/stretchr/testify) for assertions, and [gomock](https://github.com/uber-go/mock) for mock implementations. The standard [errors](https://pkg.go.dev/errors) package is used; not any third-party error packages.

Consider recommendations from these resources:

- [Effective Go](https://go.dev/doc/effective_go).
- [Go code review comments](https://github.com/golang/go/wiki/CodeReviewComments).
- [Practical Go](https://dave.cheney.net/practical-go/presentations/gophercon-singapore-2019.html).

### Node

Node code is written only in [TypeScript](https://www.typescriptlang.org/), and uses [Jest](https://jestjs.io/) as the testing framework.

[ESLint](https://typescript-eslint.io/) is used to apply linting checks. Consider using an [editor integration](https://eslint.org/docs/latest/use/integrations) to avoid linting failures.

[Prettier](https://prettier.io/) is used to apply consistent code formatting. Consider using an [editor integration](https://prettier.io/docs/en/editors) to help match Prettier's formatting.

### Java

Java code uses [JUnit](https://junit.org/) as the test runner, [AssertJ](https://assertj.github.io/doc/) for assertions, and [Mockito](https://site.mockito.org/) for mock implementations. [Checkstyle](https://checkstyle.org/) is used to apply both linting checks and consistent code formatting.

Consider recommendations from these resources:

- Effective Java by Joshua Bloch.

### Scenario tests

Scenario tests use [Cucumber](https://cucumber.io/docs/cucumber/) as the testing framework. The same set of test features are run against all language implementations of the API to ensure consistent functionality and behavior.

## Code of conduct

Please review the project [code of conduct](CODE_OF_CONDUCT.md) before contributing.

## Maintainers

Should you have any questions or concerns, please reach out to one of the project's [maintainers](MAINTAINERS.md).

## Hyperledger Fabric

See the [Hyperledger Fabric contributors guide](http://hyperledger-fabric.readthedocs.io/en/latest/CONTRIBUTING.html) for more details, including other Hyperledger Fabric projects you may wish to contribute to.

---

[![Creative Commons License](https://i.creativecommons.org/l/by/4.0/88x31.png)](http://creativecommons.org/licenses/by/4.0/)
This work is licensed under a [Creative Commons Attribution 4.0 International License](http://creativecommons.org/licenses/by/4.0/)
