## slack datastore bulk-delete

Delete multiple items from a datastore

### Synopsis

Delete multiple items from a datastore.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack datastore bulk-delete <expression> [flags]
```

### Examples

```
# Delete two items from the datastore
$ slack datastore bulk-delete --datastore tasks '{"ids": ["12", "42"]}'

# Delete two items from the datastore with an expression
$ slack datastore bulk-delete '{"datastore": "tasks", "ids": ["12", "42"]}'
```

### Options

```
      --datastore string   the datastore used to store items
  -h, --help               help for bulk-delete
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

