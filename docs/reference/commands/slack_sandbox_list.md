# `slack sandbox list`

List developer sandboxes

## Description

List details about your developer sandboxes.

The listed developer sandboxes belong to a developer program account
that matches the email address of the authenticated user.

```
slack sandbox list [flags]
```

## Flags

```
  -h, --help            help for list
      --status string   Filter by status: active, archived
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
$ slack sandbox list                  # List developer sandboxes
$ slack sandbox list --status active  # List active sandboxes only
```

## See also

* [slack sandbox](slack_sandbox)	 - Manage developer sandboxes

