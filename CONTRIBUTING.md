# Contributing to `zk`

## Understanding the codebase

### Building the project

It is recommended to use the `Makefile` for compiling the project, as the `go` command requires a few parameters.

```shell
make build
```

This will be expanded to the following command:

```shell
CGO_ENABLED=1 GOARCH=arm64 go build -tags "fts5" -ldflags "-X=main.Version=`git describe --tags --match v[0-9]* 2> /dev/null` -X=main.Build=`git rev-parse --short HEAD`"
```

- `CGO_ENABLED=1` enables CGO, which is required by the `mattn/go-sqlite3` dependency.
- `GOARCH=arm64` is only required for Apple Silicon chips.
- `-tags "fts5"` enables the FTS option with `mattn/go-sqlite3`, which handles much of the magic behind `zk`'s `--match` filtering option.
- ``-ldflags "-X=main.Version=`git describe --tags --match v[0-9]* 2> /dev/null` -X=main.Build=`git rev-parse --short HEAD`"`` will automatically set `zk`'s build and version numbers using the latest Git tag and commit SHA.

### Automated tests

The project is vetted with two different kind of automated tests: unit tests and end-to-end tests.

#### Unit tests

Unit tests are using the standard [Go testing library](https://pkg.go.dev/testing). To execute them, use the command `make test`.

They are ideal for testing parsing output or individual API edge cases and minutiae.

#### End-to-end tests

Most of `zk`'s functionality is tested with functional tests ran with [`tesh`](https://github.com/mickael-menu/tesh), which you can execute with `make tesh` (or `make teshb`, to debug whitespaces changes).

When addressing a GitHub issue, it's a good idea to begin by creating a `tesh` file in `tests/issue-XXX.tesh`. If a starting notebook state is required, it can be added under `tests/fixtures`.

If you modify the output of `zk`, you may disrupt some `tesh` files. You can use `make tesh-update` to automatically update them with the correct output.

### CI workflows

Several GitHub action workflows are executed when pull requests are merged or releases are created.

- `.github/workflows/build.yml` checks that the project can be built and the tests still pass.
- `.github/workflows/codeql.yml` runs static analysis to vet code quality.
- `.github/workflows/gh-pages.yml` deploy the documentation files to GitHub Pages.
- `.github/workflows/release.yml` submits a new version to Homebrew when a Git version tag is created.
- `.github/workflows/triage.yml` automatically tags old issues and PRs as staled.

## Releasing a new version

When `zk` is ready to be released, you can update the `CHANGELOG.md` ([for example](https://github.com/zk-org/zk/commit/ea4457ad671aa85a6b15747460c6f2c9ad61bf73)) and create a new Git version tag (for example `v0.13.0`). Make sure you follow the [Semantic Versioning](https://semver.org) scheme.

Then, create [a new GitHub release](https://github.com/zk-org/zk/releases) with a copy of the latest `CHANGELOG.md` entries and the binaries for all supported platforms.

Binaries can be created automatically using `make dist-linux` and `make dist-macos`.

Unfortunately, `make dist-macos` must be run manually on both an Apple Silicon and Intel chips. The Linux builds are created using Docker and [these custom images](https://github.com/zk-org/zk-xcompile), which are hosted via [ghcr.io within zk-org](https://github.com/orgs/zk-org/packages/container/package/zk-xcompile).

This process is convoluted because `zk` requires CGO with `mattn/go-sqlite3`, which prevents using Go's native cross-compilation. Transitioning to a CGO-free SQLite driver such as [cznic/sqlite](https://gitlab.com/cznic/sqlite) could simplify the distribution process significantly.

## Documentation

TODO: add documentation steps for Sphinx docs, after it's all working.
