# `slack sandbox`

Manage developer sandboxes

## Description

Manage Slack developer sandboxes without leaving your terminal.
Use the --team flag to select the authentication to use for these commands.

Prefer a UI? Head over to
[https://api.slack.com/developer-program/sandboxes](https://api.slack.com/developer-program/sandboxes)

New to the Developer Program? Sign up at
[https://api.slack.com/developer-program/join](https://api.slack.com/developer-program/join)

```
slack sandbox <subcommand> [flags] --experiment=sandboxes
```

## Flags

```
  -h, --help   help for sandbox
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

## See also

* [slack](slack)	 - Slack command-line tool
* [slack sandbox list](slack_sandbox_list)	 - List developer sandboxes

