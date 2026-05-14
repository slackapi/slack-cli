# `slack version`

Print the version number

## Description

All software has versions. This is ours.

Version numbers follow the semantic versioning specification (semver)
and are always prefixed with a `v`, such as `v3.0.1`.

Given a version number MAJOR.MINOR.PATCH:

1. MAJOR versions have incompatible, breaking changes
2. MINOR versions add functionality that is a backward compatible
3. PATCH versions make bug fixes that are backward compatible

Experiments are patch version until officially released.

Development builds use `git describe` and contain helpful info,
such as the prior release and specific commit of the build.

Given a version number `v3.0.1-7-g822d09a`:

1. `v3.0.1`   is the version of the prior release
2. `7`        is the number of commits ahead of the `v3.0.1` git tag
3. `g822d09a` is the git commit for this build, prefixed with `g`

```
slack version [flags]
```

## Flags

```
  -h, --help   help for version
```

## Global flags

```
  -a, --app string           use a specific app ID or environment
      --config-dir string    use a custom path for system config directory
  -e, --experiment strings   use the experiment(s) in the command
  -f, --force                ignore warnings and continue executing command
      --no-color             remove styles and formatting from outputs
  -s, --skip-update          skip checking for latest version of CLI
  -w, --team string          select workspace or organization by team name or ID
      --token string         set the access token associated with a team
  -v, --verbose              print debug logging and additional info
```

## Examples

```
$ slack version                  # Print version number using the command
$ slack --version                # Print version number using the flag
$ slack --version --skip-update  # Print version and skip update check
```

## See also

* [slack](slack)	 - Slack command-line tool

