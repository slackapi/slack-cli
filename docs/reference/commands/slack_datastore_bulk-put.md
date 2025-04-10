## slack datastore bulk-put

Create or replace a list of items in a datastore

### Synopsis

Create or replace a list of items in a datastore.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack datastore bulk-put <expression> [flags]
```

### Examples

```
# Create or replace two new entries in the datastore
$ slack datastore bulk-put --datastore tasks '{"items": [{"id": "12", "description": "Create a PR", "status": "Done"}, {"id": "42", "description": "Approve a PR", "status": "Pending"}]}'

# Create or replace two new entries in the datastore with an expression
$ slack datastore bulk-put '{"datastore": "tasks", "items": [{"id": "12", "description": "Create a PR", "status": "Done"}, {"id": "42", "description": "Approve a PR", "status": "Pending"}]}'
```

### Options

```
      --datastore string   the datastore used to store items
      --from-file string   store multiple items from a file of JSON Lines
  -h, --help               help for bulk-put
      --show               only construct a JSON expression
      --unstable           kick the tires of experimental features
```

### Options inherited from parent commands

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

### SEE ALSO

* [slack datastore](slack_datastore)	 - Interact with an app's datastore

