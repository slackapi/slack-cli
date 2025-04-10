## slack datastore delete

Delete an item from a datastore

### Synopsis

Delete an item from a datastore.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack datastore delete <expression> [flags]
```

### Examples

```
# Remove an item from the datastore
$ slack datastore delete --datastore tasks '{"id": "42"}'

# Remove an item from the datastore with an expression
$ slack datastore delete '{"datastore": "tasks", "id": "42"}'
```

### Options

```
      --datastore string   the datastore used to store items
  -h, --help               help for delete
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

