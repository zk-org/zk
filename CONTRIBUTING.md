# Contributing to `zk`

## Understanding the codebase

We have a `dev` and `main` branch. If you are contributing, then branch from
`dev`.

### Building the project

It is recommended to use the `Makefile` for compiling the project, as the `go`
command requires a few parameters.

```shell
make build
```

This will be expanded to the following command:

```shell
CGO_ENABLED=1 GOARCH=arm64 go build -tags "fts5" -ldflags "-X=main.Version=`git describe --tags --match v[0-9]* 2> /dev/null` -X=main.Build=`git rev-parse --short HEAD`"
```

- `CGO_ENABLED=1` enables CGO, which is required by the `mattn/go-sqlite3`
  dependency.
- `GOARCH=arm64` is only required for Apple Silicon chips.
- `-tags "fts5"` enables the FTS option with `mattn/go-sqlite3`, which handles
  much of the magic behind `zk`'s `--match` filtering option.
- ``-ldflags "-X=main.Version=`git describe --tags --match v[0-9]* 2> /dev/null`"``
  will automatically set `zk`'s build and version numbers using the latest Git
  tag and commit SHA.

### Automated tests

The project is vetted with two different kind of automated tests: unit tests and
end-to-end tests.

#### Unit tests

Unit tests are using the standard
[Go testing library](https://pkg.go.dev/testing). To execute them, use the
command `make test`.

They are ideal for testing parsing output or individual API edge cases and
minutiae.

#### End-to-end tests

Most of `zk`'s functionality is tested with functional tests ran with
[`tesh`](https://github.com/mickael-menu/tesh), which you can execute with
`make tesh` (or `make teshb`, to debug whitespaces changes).

When addressing a GitHub issue, it's a good idea to begin by creating a `tesh`
file in `tests/issue-XXX.tesh`. If a starting notebook state is required, it can
be added under `tests/fixtures`.

If you modify the output of `zk`, you may disrupt some `tesh` files. You can use
`make tesh-update` to automatically update them with the correct output.

### CI workflows

Several GitHub action workflows are executed when pull requests are merged or
releases are created.

- `.github/workflows/build.yml` checks that the project can be built and the
  tests still pass.
- `.github/workflows/build-binaries.yml` builds zk binaries for all platforms
  and uploads them.
- `.github/workflows/codeql.yml` runs static analysis to vet code quality.
- `.github/workflows/build-docs.yml` builds the docs site.
- `.github/workflows/gh-pages.yml` deploys the documentation files to GitHub
  Pages.
- `.github/workflows/release.yml` runs the `build-binaries` workflow, downloads
  its artifacts and then attaches them to a new draft release.
- `.github/workflows/triage.yml` automatically tags old issues and PRs as
  staled.

## Documentation

We're using [Sphinx](https://www.sphinx-doc.org/en/master/) as our documentation
framework, and the [furo](https://pradyunsg.me/furo/quickstart/) theme.

To install, from the repository root run:

```sh
pip install -r docs/requirements.txt
```

`docs/` is the root level of the documentation site.
[index.rst](./docs/index.rst) is the landing page.

Documentation is written in markdown, with the exception of pages which render
TOCs. These pages are written in
[reStructuredText (rst)](https://www.sphinx-doc.org/en/master/usage/restructuredtext/basics.html),
as Myst (which does the parsing from Markdown to rst within Sphinx's back end),
does not handle TOCs yet (as far as I'm (@tjex) aware).

### Formatting

There is a `.prettierrc` at the root of the git repo. If you are using a
different formatter, feel free to add its config file to this repo with the same
settings.

### Local Preview

Sphinx generates static html. So previewing locally is easy. Simply build the
site with `make`:

```sh
make zkdocs
```

This will create a folder `[docs-build/]` containing the static site (and is of
course ignored by git, so you can do whatever you like in that folder).

Open `docs-build/index.html` in your browser and you're good to go.

You can install and use
[sphinx-autobuild](https://pypi.org/project/sphinx-autobuild/) to emulate hot
reloading / live server style development.\
Otherwise you can just manually rebuild with `make zkdocs` each time you want to
preview your changes.

### Deploying

Deployment to the world wide web happens via GitHub actions upon a PR to the
main branch.

So commit and push your changes to your own branch, and make a PR as usual.\
Once merged to main, the site will be build and deployed.

## Releasing a new version

When `zk` is ready to be released, follow these steps in order:

1. Update the `CHANGELOG.md`
   ([for example](https://github.com/zk-org/zk/commit/ea4457ad671aa85a6b15747460c6f2c9ad61bf73)).
2. Commit the changes above with `git commit` (no `-m`). In the first line of
   the commit, provide "Release <the-version>". List any necessary detail on
   subsequent lines.
3. Finally, create a new Git version tag with `git tag -a <version>`(syntax
   example: `v0.13.0`). Make sure you follow the
   [Semantic Versioning](https://semver.org) scheme.

If you create the git tag via the command line, and push it (`git push --tags`), then the
[release action](.github/workflows/release.yml) will be triggered. This in turn
calls the [build-binaries action](.github/workflows/build-binaries.yml), creates
a _draft_ release on GitHub and attaches the built binaries.

Alternatively, you can manually create a release via the GitHub interface, also
creating a release tag. Then you would run the
[build-binaries action](.github/workflows/build-binaries.yml) manually, and
again manually download and attach the binaries.

In both cases the description of the release can be edited after the release is
created (i.e, adding or editing the changelog).

