# Maintainers' Guide

This document describes the tools, tasks, and workflows that one needs to be
familiar with in order to effectively maintain this project. If you use this
project for your own development as is but don't plan on modifying it, this
guide is **not** for you. You can of course keep reading though.

**Outline of the guide:**

- [Tools](#tools): the development tooling for this codebase
- [Project layout](#project-layout): an overview of project directories
- [Tasks](#tasks): common things done during development
- [Development build](#development-build): for releases with the latest changes
- [Workflow](#workflow): around changes and contributions
- [Everything else](#everything-else): and all of those other things

## Tools

### Golang

Most development is done with the Go programming language.

To get started as a maintainer, you'll need to install [Golang][golang]. We tend
to use the latest version available but our minimum version is defined in the
[`go.mod`][gomod] file.

We recommend using [the official installer][goinstaller] to install the matching
version.

#### Troubleshooting

Sometimes setups break but please don't fear. Solutions remain in reach.

<details>
<summary>Finding and setting the GOPATH</summary>

The `GOPATH` is an environment variable that determines where `go` checks for
certain files, so correct values are needed to run certain `Makefile` commands.

If packages for this project fail to be found during setup, checking these might
be a good place to start!

First, find your `GOPATH`. The easiest way to run the following commands:

```sh
go env | grep GOPATH
```

Next, update your `GOPATH` environment variables in your `~/.zshrc`
([learn how for other terminals like bash][gopath]) to point to your
installation:

```sh
# Set variables in .zshrc file

# NOTE - your `GOPATH` may be different
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

</details>

<details>
<summary>Updating the GOOS and GOARCH</summary>

Compiling the project might work just fine but an attempt to run the build can
error with something like an `exec format error`.

This happens if you're compiling for the wrong target machine and can be fixed
by adjusting your `GOOS` and `GOARCH` environment variables to match the current
system you're using.

</details>

<details>
<summary>Uninstalling a Homebrew installation</summary>

We don't recommend installing Golang with [Homebrew][homebrew] because it can be
difficult to choose a specific version and set the proper paths.

Uninstall Golang from Homebrew:

```sh
brew uninstall go
```

</details>

### golangci-lint

Linters are used to help catch formatting strangeness and strange spellings.

The CI runs linting via [golangci-lint][golangci-lint] and is configured in the
`.golangci.yml` file.

Install this tool to mimic CI environments and run linting with the `Makefile`:

```sh
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

Confirm the installation worked right with:

```sh
golangci-lint --version
```

#### VSCode

Settings are already defined in this project for using the VSCode editor with
`golangci-lint` as the `go.lintTool`. Inspect [`.vscode/settings.json`][vscode]
for more details.

#### Troubleshooting

Linting troubles can happen from unexpected formatting or a broken installation.
Broken installations might be fixed by the below:

<details>
<summary>Matching versions that don't match</summary>

The latest version of `golangci-lint` is used in testing and should be matched
by the installation you have.

Compare [the latest version][golangci-lint-release] to your current installation
and try linting once more.

</details>

<details>
<summary>Uninstalling a Homebrew installation</summary>

Setting the correct version of `golangci-lint` can be difficult with Homebrew.
If this is causing issues try uninstalling it with:

```sh
brew uninstall golangci-lint
```

</details>

### Makefile

Common commands are contained in the `Makefile` for this project. These perform
tasks such as building and testing and releasing.

More details will follow these different tasks but you should have `make` ready:

```sh
make --version
```

## Project layout

Our project layout is based on [golang-standards][golang-standards],
[Effective Go][effective-go], and [Practical Go][practical-go].

```txt
bin/ ........ build artifacts including the compiled CLI binary
cmd/ ........ commands some logic to process inputs and outputs
dist/ ....... release artifact that outputs the signed binaries
internal/ ... private packages that contain shared logic for `cmd`
scripts/ .... files for downloading and installing released builds
test/ ....... test data and test helper packages
```

### `bin/`

The build artifact from `make build` that outputs the compiled CLI binary.

### `cmd/`

Defines [each command that is available][commands] to the command-line tool.

Think of this as the user's interface into this command-line tool. Each command
processes input from `stdin` then outputs information to `stdout` and `stderr`
as certain actions are performed.

The control flow of a command is contained with that command, while some details
and execution implementations are delegated to `pkg/`.

The file structure maps directly to the command, such as:

```txt
$ slack auth login  # full command
$ slack login       # shortcut command

cmd/
|__ auth/
    |___ login.go
```

### `dist/`

The release build artifacts from `make build-snapshot` that includes binaries
and archives for all platforms.

An installation of [`goreleaser`][goreleaser] is required and the latest version
is recommended.

### `internal/`

A safe place to store private packages used by commands and this program. Shared
logic and implementation specifics are contained within since these packages are
not exported.

#### `internal/pkg/`

Some detailed command implementations extend into these internal packages when
the implementations are detailed. This is meant for command logic that is needed
to perform the command, but might not make sense to inline for some reason such
as testing.

Think of this as the back-end for a command. Each package should implement logic
for a single command, but delegate input and outputs to the command front-end.

The file structure maps directly to the command and should usually match `cmd/`,
such as:

```txt
$ slack auth login  # full command
$ slack login       # shortcut command

cmd/
|__ auth/
|   |__ login.go
pkg/
|__ auth/
    |__ login.go
```

### `scripts/`

Installation and setup scripts for various operating systems are available here.

### `test/`

This is home to test data, helpers, and anything else test related. The only
exception are the unit test files, which sit next to the file they test with the
suffix `_test.go`.

## Tasks

Certain things are common during development and require a few commands.

**Task outline:**

- [Cloning the project](#cloning-the-project)
- [Initializing the project](#initializing-the-project)
- [Testing](#testing)
  - [Module and unit tests](#module-and-unit-tests)
  - [End-to-end tests](#end-to-end-tests)
  - [Continuous integration tests](#continuous-integration-tests)
  - [Contributing tests](#contributing-tests)
- [Building](#building)
- [Running](#running)
  - [Setting a global alias](#setting-a-global-alias)
- [Generating documentation](#generating-documentation)
  - [Synchronizing changes upstream](#synchronizing-changes-upstream)
- [Versioning](#versioning)
- [Updating](#updating)
  - [Bumping the Golang version](#bumping-the-golang-version)
  - [Bumping Go packages versions](#bumping-go-package-versions)
  - [Bumping Goreleaser versions](#bumping-goreleaser-versions)
- [Deprecating features and flags](#deprecating-features-and-flags)
- [Allowlist configuration](#allowlist-configuration)

### Cloning the project

```zsh
git clone https://github.com/slackapi/slack-cli.git
cd slack-cli/
```

### Initializing the project

When you first clone the project or after you major updates to the project, it's
a good idea to run our initialization script. It will fetch the latest
dependencies, tags, and anything else to keep you up-to-date:

```zsh
make init
```

### Testing

To ensure that certain code paths are working as we might expect, we use a suite
of module-based unit tests and complete end-to-end tests, along with a healthy
dose of manual testing and dogfooding.

#### Module and unit tests

Each module has unit tests in the same directory as the code with the suffix
`_test.go` (i.e. `version_test.go`).

You can run the entire unit test suite using the command:

```zsh
make test
```

You can also add the `testdir` argument to specify which directories to test and
`testname` to pick the name of tests to run:

```zsh
make test testdir=cmd/auth testname=TestLogoutCommand
```

Note: Test code should be written in syntax that runs on the oldest supported Go
version defined in
[`go.mod`](https://github.com/slackapi/slack-cli/blob/main/go.mod). This ensures
that backwards compatibility is tested and the APIs work in versions of Golang
that do not support the most modern packages.

To get a visual overview of unit test coverage on a per-file basis:

```zsh
make test      # run the test suite
make coverage  # open coverage summary
```

#### End-to-end tests

A suite of end-to-end tests that cover common commands with the CLI, samples,
and the client can be found in the private
[slackapi/platform-devxp-test][platform-devxp-test] repo. Please see that repo
for details on how to run the end-to-end tests locally.

These tests take about 5 minutes to complete and the test results will be reported
on CircleCI. At this time only members of the `slackapi` GitHub org can access test
results directly.

#### Continuous integration tests

All tests are run during PRs, merges to `main`, and on nightly builds of `main`
by our continuous integration systems: GitHub Action and CircleCI workflows.

The module tests that are used can be found alongside the modules in `*_test.go`
files and edited as needed.

The end-to-end tests are located in the
[slackapi/platform-devxp-test][platform-devxp-test] repo and can be adjusted to
new requirements by a maintainer.

By default the tests execute with the `dev-build` GitHub Release of this repo.
Branches on this repo will also create a related GitHub Release, and end-to-end
tests will use the branch-specific Slack CLI release to execute tests.

The branch name can also be set by changing
[the `e2e_target_branch`](https://github.com/slackapi/slack-cli/blob/24048e34b30a1f1bed46f9d937ff0b02a3cb76b7/.circleci/config.yml#L950)
for the `build-lint-test-e2e-test` workflow in the `.circleci/config.yml` file,
but take care not to merge this change into `main`!

#### Contributing tests

If you'd like to add tests, please review our
[contributing guide](./CONTRIBUTING.md) to learn how to open a discussion and
pull request. We'd love to talk through setting up tests!

Updating tests in [the end-to-end testing repo][platform-devxp-test] can be done
by a maintainer through opening a pull request to that repo with the changes.

### Building

This program is compiled into a single, executable binary. Before building, the
build script will remove past artifacts with `make clean`:

```zsh
make build
```

### Running

You can run the executable binary with:

```zsh
./bin/slack --version
```

If you have installed the official release of the CLI, then you will now have
two binaries:

```zsh
slack --version        # this is the official release
./bin/slack --version  # this is the local development build
```

#### Setting a global alias

Some developers like to use the local development build as their globally
installed binary, so that they do not need to prefix the binary with `./bin/`.
You can do this in a couple of different ways:

- [Option 1: Replace the global binary](#option-1-replace-the-global-binary)
- [Option 2: Symlink the compiled binary](#option-2-symlink-the-compiled-binary)

##### Option 1: Replace the global binary

If `/usr/local/bin/slack` already points a different binary, you can choose to
update it globally.

```zsh
which slack   # Check the path for slack
slack         # Check what program executes when you run slack
```

Delete the official release from `/usr/local/bin/slack` if you installed it in
the past:

```zsh
rm /usr/local/bin/slack
```

Next, update your `$PATH` to include your local build by opening your `~/.zshrc`
or `~/.bash_profile` and adding the following:

```zsh
# Replace `path/to/slack-cli` with your project's full path
# You can find the path by running `pwd` inside the slack-cli/ directory
export PATH=path/to/slack-cli/bin:$PATH
```

Then, refresh your terminal session:

```zsh
source ~/.zshrc
# or, source ~/.bash_profile for bash terminals
# or, just close and re-open your terminal
```

Finally, verify that you have correctly set up your `$PATH` by running the CLI:

```zsh
slack --version
```

##### Option 2: Symlink the compiled binary

Create a symbolic link to a build of this program with the following command and
changes:

- Replace `/path/to/` with your path
- Replace `slack-cli-local` with your preferred name to reference the compiled
  binary

```zsh
ln -s ~/path/to/slack-cli/bin/slack /usr/local/bin/slack-cli-local
```

Verify that everything is linked correctly with a test in the home directory:

```zsh
cd
slack-cli-local --version
```

### Generating documentation

You can generate documentation for all commands in the `docs/` directory with:

```zsh
slack docgen
```

#### Synchronizing changes upstream

Automated workflows run on documentation changes to [sync][sync] files between
this project and the documentation build.

A GitHub application called [`@slackapi[bot]`][github-app-docs] mirrors changes
to these files and requires certain permissions:

- **Actions**: Read and write
- **Contents**: Read and write
- **Metadata**: Read
- **Pull requests**: Read and write

Access to both this project repo and documentation repo must also be granted.

Credentials and secrets for the app can be stored as the following variables:

- `GH_APP_ID_DOCS`
- `GH_APP_PRIVATE_KEY_DOCS`

### Versioning

We use git tags and [semantic versioning][semver] to version this program. The
format of a tag should always be prefixed with `v`, such as `v0.14.0`.

Changes behind an experiment flag should be `semver:patch`. This is because
experimental features don't impact regular developers and experiments are
not documented. Once the experimental flag is removed, a `semver:patch`,
`semver:minor`, or `semver:major` version should be tagged.

The version is generated using `git describe` and contains some helpful
information:

```zsh
./bin/slack-cli-local --version
slack-cli-local v0.14.0-7-g822d09a

# v0.14.0  is the git tag release version
# 7        is the number of commits this build is ahead of the v0.14.0 tag
# g822d09a is the git commit for this build
```

Pushing a new version with `git tag` and a semantic version creates a release
for download with the installation scripts. Check out
[Development Build](#development-build).

Internal runbooks detail the process for official releases of stable versions.

### Updating

The code this project relies upon have occasional updates that can be pulled in
with some processes.

#### Bumping the Golang version

New versions of the Go language are [released with great features][goreleases]
that can be useful in development. Updates for the latest are checked every day
for both the latest **minor** or **patch** release.

A pull request is created once a newer version is found, attempting to update
the specific references language references we have. Human checks should happen
to ensure all references were updated properly:

- The Go mod file in: `go.mod` - e.g. `go 1.22.1`
- CircleCI runners in: `.circleci/config.yml` - e.g. `cimg/go:1.22.1`
- GitHub Actions in: `.github/workflows/tests.yml` - e.g. `actions/setup-go`

Automation that powers can be found in [this workflow][wf-dependencies] and
[this app][github-app-releaser]. Secrets are found elsewhere.

For these changes to complete, certain application permissions are needed:

- **Actions**: Read and write
- **Contents**: Read and write
- **Metadata**: Read
- **Pull requests**: Read and write
- **Workflows**: Read and write

Access to this project is also required with the selected application scopes.

Credentials and secrets for the app can be stored as the following variables:

- `GH_APP_ID_RELEASER`
- `GH_APP_PRIVATE_KEY_RELEASER`

#### Bumping Go package versions

Dependencies are often updated with new and cool things too. Most of these are
found with the help of @dependabot but more updates remain possible with:

- Update the module version in `go.mod`
- Run `go mod tidy` to update the modules and the `go.sum`

To see the version of a dependency this module uses, or to see the dependency
tree of a transitive dependency, this command can be helpful:

```zsh
go mod graph | grep <module name>
```

#### Bumping Goreleaser versions

The [`goreleaser`][goreleaser] package we use to build release snapshots needs
updates in the following files on occasion:

- `.circleci/config.yml`
- `.goreleaser-dev.yml`
- `.goreleaser.yml`

Testing in our CI setup uses changes to these files when creating test builds.

### Deprecating features and flags

Many good things come to an end. This can sometimes include commands and flags.
When commands or flags need to be removed, follow these steps:

<details>
<summary>Deprecating features</summary>

- Public functionality should be deprecated on the next `semver:major` version
  - Add the comment `// DEPRECATED(semver:major): Description about the deprecation and migration path`
  - Print a warning `PrintWarning("DEPRECATED: Description about the deprecation and migration path")`
- Internal functionality can be deprecated anytime
  - Add the comment `// DEPRECATED: Description about the deprecation and migration path`
- Please add deprecation comments generously to help the person completely remove the feature and tests

</details>

<details>
<summary>Deprecating commands</summary>

- Display a deprecation notice on the command with the `Deprecated` attribute
- Add a deprecation warning to the `Long` help message
- Optionally hide the command from help menus with the `Hidden` attribute
- Somehow mark the command for removal in the next major version

</details>

<details>
<summary>Deprecating flags</summary>

- Print a message that says the flag is deprecated whenever it is used
- Recommend alternative flags if available (ex: `internal/config/flags.go`)
- Hide the flag from help menus with the `.Hidden` attribute
- Optionally use the recommended flag whenever possible
- Somehow mark the command for removal in the next major version

</details>

### Allowlist configuration

The following inbound/outbound domains are required for Slack CLI to function
properly:

Slack CLI:

- api.slack.com
- downloads.slack-edge.com
- slack.com
- slackb.com

## Development build

The development build comes in 2 flavours:

1. Development build GitHub release
2. Development build install script

### 1. Development build GitHub release

A development build and recent changelog is generated each night from `main`
with all of the latest changes. Builds are released with the `dev-build` tag and
can be [reviewed from the releases][dev-release] page.

Each release page contains:

- The commit used when building the development build and evidence of a tag
  created in past versions
- A changelog containing commit messages made since the last tag with a semantic
  version `v*.*.*` pattern
- Assets for the macOS, Linux, and Windows binaries

The development build and release automation is performed from the `deploy-dev`
job in the [`.circleci/config.yml`][circleci] file.

### 2. Development build install script

An installation script for the development build provides the same `dev-build`
release tag but with magic setup:

```bash
curl -fsSL https://downloads.slack-edge.com/slack-cli/install-dev.sh | bash -s slack-cli-dev
```

Changes to the actual installation scripts are made through other channels and
might become outdated between releases. These scripts can still be found in the
[`scripts/`][scripts] directory in the meantime.

## Workflow

### Fork

Development on a forked branch is common for contributors to this project and
encouraged for submitting contributions. Read [`CONTRIBUTING.md`][contributing]
for instructions on contributing.

Maintainers might prefer to push to the origin repostiory for access to secret
environment variables present in CI workflows. This is fine but with care taken
to use **branches**.

### Branches

`main` is where active development occurs.

Development should happen on branches created off of `main` or other branches
and namespaced according to the change being made to help remain organized.

A [conventional commit][commit] pattern is recommended to help group branches
that are making changes to similar parts of the codebase - e.g.
`fix-config-auth-id-missing` or `feat-command-help-interactive`.

Sharing branch names can be useful for collaboration on long-running features
and is recommended for pulling changes from others without pushing to the same
remote or pull request. Otherwise branch naming is unique to the change being
made.

After a major version increment, there also may be maintenance branches created
specifically for supporting older major versions.

### Experimental features

Curious or promising feature work might prefer to remain behind an experiment
flag to avoid changing current command behaviors while still merging into the
`main` branch.

Experiments are introduced in `internal/shared/experiment/experiment.go` and
follow the kebab-case-format, while implementation of features can be gated with
the `config.WithExperimentOn` utility.

Run `slack --help --verbose` to view all active experiments.

Learn more about how experiments are versioned in the [Versioning section](#versioning).

### Issue management

Labels are used to run issues through an organized workflow. Here are the basic
definitions:

- `bug`: A confirmed bug report. A bug is considered confirmed when reproduction
  steps have been documented and the issue has been reproduced.
- `build`: A change to the build, compile, or CI/CD pipeline.
- `changelog`: An issue or pull request that should be mentioned in the public
  release notes or CHANGELOG.
- `code health`: An issue or pull request that is focused on internal refactors
  or tests.
- `discussion`: An issue that is purely meant to hold a discussion. Typically
  the maintainers are looking for feedback in this issues.
- `docs`: An issue that is purely about documentation work.
- `duplicate`: An issue that is functionally the same as another issue. Apply
  this only if you've linked the other issue by number.
- `enhancement`: A feature request for something this package might not already
  do.
- `experiment`: A change that is accessed behind the --experiment flag or toggle
- `good first issue`: An issue that has a well-defined relatively-small scope,
  with clear expectations. It helps when the testing approach is also known.
- `needs info`: An issue that may have claimed to be a bug but was not
  reproducible, or was otherwise missing some information.
- `question`: An issue that is like a support request because the user's usage
  was not correct.
- `security`: An issue that has special consideration for security reasons.
- `semver:major`: A change that requires a semver major release.
- `semver:minor`: A change that requires a semver minor release.
- `semver:patch`: A change that requires a semver patch release.
- `server side issue`: An issue that exists on the Slack Platform, Slack API,
  or other remote endpoint.

**Triage** is the process of taking new issues that aren't yet "seen" and
marking them with a basic level of information with labels. An issue should have
**one** of the following labels applied:

- `bug`
- `build`
- `code health`
- `discussion`
- `docs`
- `enhancement`
- `needs info`
- `question`

_Hint: The main triage issues always have a description that starts with `M-T:`_

Issues are closed when a resolution has been reached or the issue becomes stale.
If for any reason a closed issue seems relevant once again, reopening is great
and better than creating a duplicate issue.

### Pull requests

#### Pull request: triage

Opened pull requests should be triaged by a maintainer. The purpose of pull
request triage is to keep the project healthy, organized, and to help simplify
releases. The expectation is that maintainers triage their own pull requests and
those from outside contributors. These are transferrable practices that are
followed by most of Slack's open sourced SDKs and Frameworks.

Steps to triage a pull request:

1. **Reviewers**:
   - Minimum of 1 reviewer
2. **Assignees**:
   - Minimum of 1 assignee
   - Assignees are usually the people who are committing changes
3. **Labels**:
   1. One of the following main labels to describe the type of pull request:
      - Main labels always have a description that starts with `M-T:`
        - Example: `enhancement`, `bug`, `discussion`, `documentation`, `code health`
        - Note: The main labels are used to organize the automated CHANGELOG
      - The `changelog` label denotes changes to document in the public Slack
        API Docs release notes
      - Learn about labels in the [Issue management section](#issue-management)
   2. A semver label
      - Semver labels are used to determine the next release's version
      - Example: `semver:major`, `semver:minor`, or `semver:patch`
      - Learn more about semver in the [Versioning section](#versioning)
4. **Milestone**:
   - A milestone should be assigned when possible, usually as the `Next Release`
   - After a release, the `Next Release` milestone is renamed to the tagged
     version

#### Pull request: merge

Steps to merge a pull request:

1. Tests must pass on the pull request using the continuous integration suite
   - Tests for development APIs are optional, but we recommend investigating why
     it's failing before merging
   - End-to-end tests for pull requests from forks can be [started][e2e] if:
     - The `workflow` from the **main** branch is used
     - The `branch` contains the pull request number as: `pull/123/head`
     - The `status` is reported with the commit checks
2. Code is reviewed and approved by another maintainer
3. Title is descriptive
   - This becomes part of the CHANGELOG, so please make sure it's meaningful to
     developers
   - References original issue when possible
   - Follows the [conventional commit][commit] format
4. Squash and merge the pull request
   - Anyone can merge the pull request, but usually it's the creator, assignees,
     or reviewers
   - Commit message should reference original issue and pull request in commit

## Everything else

When in doubt, find the other maintainers and ask.

[circleci]: ../.circleci/config.yml
[commands]: https://tools.slack.dev/slack-cli/reference/commands/slack
[commit]: https://www.conventionalcommits.org/en/v1.0.0/
[contributing]: ./CONTRIBUTING.md
[dev-release]: https://github.com/slackapi/slack-cli/releases/tag/dev-build
[e2e]: https://github.com/slackapi/slack-cli/actions/workflows/e2e_tests.yml
[effective-go]: https://golang.org/doc/effective_go
[github-app-docs]: https://github.com/apps/slackapi
[github-app-releaser]: https://github.com/apps/slack-cli-releaser
[goinstaller]: https://go.dev/doc/install
[golang]: https://golang.org/
[golang-standards]: https://github.com/golang-standards/project-layout
[golangci-lint]: https://golangci-lint.run/
[golangci-lint-release]: https://github.com/golangci/golangci-lint/releases
[gomod]: https://github.com/slackapi/slack-cli/blob/main/go.mod
[gopath]: https://gist.github.com/vsouza/77e6b20520d07652ed7d
[goreleaser]: https://github.com/goreleaser/goreleaser
[goreleases]: https://go.dev/doc/devel/release
[goversions]: https://go.dev/dl/
[homebrew]: https://formulae.brew.sh/formula/go
[platform-devxp-test]: https://github.com/slackapi/platform-devxp-test
[practical-go]: https://dave.cheney.net/practical-go/presentations/qcon-china.html
[scripts]: ../scripts
[semver]: https://semver.org/
[sync]: https://github.com/slackapi/slack-cli/blob/main/.github/workflows/sync-docs-from-cli-repo.yml
[vscode]: https://github.com/slackapi/slack-cli/blob/main/.vscode/settings.json
[wf-dependencies]: ./workflows/dependencies.yml
