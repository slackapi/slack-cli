# `slack sandbox delete`

Delete a developer sandbox

## Description

Permanently delete a sandbox and all of its data

```
slack sandbox delete [flags]
```

## Flags

```
      --force               Skip confirmation prompt
  -h, --help                help for delete
      --sandbox-id string   Sandbox team ID to delete
```

## Global flags

```
  -a, --app string           use a specific app ID or environment
      --config-dir string    use a custom path for system config directory
  -e, --experiment strings   use the experiment(s) in the command
      --no-color             remove styles and formatting from outputs
  -s, --skip-update          skip checking for latest version of CLI
  -w, --team string          select workspace or organization by team name or ID
      --token string         set the access token associated with a team
  -v, --verbose              print debug logging and additional info
```

## Examples

```
# Delete a sandbox identified by its team ID
$ slack sandbox delete --sandbox-id E0123456
```

## See also

* [slack sandbox](slack_sandbox)	 - Manage developer sandboxes

