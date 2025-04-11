## slack datastore query

Query a datastore for items

### Synopsis

Query a datastore for items.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack datastore query <expression> [flags]
```

### Examples

```
# Collect a limited set of items from the datastore
$ slack datastore query --datastore tasks '{"limit": 8}' --output json

# Collect items from the datastore starting at a cursor
$ slack datastore query --datastore tasks '{"cursor": "eyJfX2NWaV..."}'

# Query the datastore for specific items
$ slack datastore query --datastore tasks '{"expression": "#status = :status", "expression_attributes": {"#status": "status"}, "expression_values": {":status": "In Progress"}}'

# Query the datastore for specific items with only an expression
$ slack datastore query '{"datastore": "tasks", "expression": "#status = :status", "expression_attributes": {"#status": "status"}, "expression_values": {":status": "In Progress"}}'
```

### Options

```
      --datastore string   the datastore used to store items
  -h, --help               help for query
      --output string      output format: text, json (default "text")
      --show               only construct a JSON expression
      --to-file string     save items directly to a file as JSON Lines
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

