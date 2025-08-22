# `slack datastore put`

Create or replace an item in a datastore

## Description

Create or replace an item in a datastore.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack datastore put <expression> [flags]
```

## Flags

```
      --datastore string   the datastore used to store items
  -h, --help               help for put
      --show               only construct a JSON expression
      --unstable           kick the tires of experimental features
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
# Add a new entry to the datastore
$ slack datastore put --datastore tasks '{"item": {"id": "42", "description": "Create a PR", "status": "Done"}}'

# Add a new entry to the datastore with an expression
$ slack datastore put '{"datastore": "tasks", "item": {"id": "42", "description": "Create a PR", "status": "Done"}}'
```

## See also

* [slack datastore](slack_datastore)	 - Interact with an app's datastore

