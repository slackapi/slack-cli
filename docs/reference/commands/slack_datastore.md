## slack datastore

Interact with an app's datastore

### Synopsis

Interact with the items stored in an app's datastore.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

Discover the datastores: [https://tools.slack.dev/deno-slack-sdk/guides/using-datastores](https://tools.slack.dev/deno-slack-sdk/guides/using-datastores)

```
slack datastore <subcommand> <expression> [flags]
```

### Examples

```
# Add a new entry to the datastore
$ slack datastore put --datastore tasks '{"item": {"id": "42", "description": "Create a PR", "status": "Done"}}'

# Add two new entries to the datastore
$ slack datastore bulk-put --datastore tasks '{"items": [{"id": "12", "description": "Create a PR", "status": "Done"}, {"id": "42", "description": "Approve a PR", "status": "Pending"}]}'

# Update the entry in the datastore
$ slack datastore update --datastore tasks '{"item": {"id": "42", "description": "Create a PR", "status": "Done"}}'

# Get an item from the datastore
$ slack datastore get --datastore tasks '{"id": "42"}'

# Get two items from datastore
$ slack datastore bulk-get --datastore tasks '{"ids": ["12", "42"]}'

# Remove an item from the datastore
$ slack datastore delete --datastore tasks '{"id": "42"}'

# Remove two items from the datastore
$ slack datastore bulk-delete --datastore tasks '{"ids": ["12", "42"]}'

# Query the datastore for specific items
$ slack datastore query --datastore tasks '{"expression": "#status = :status", "expression_attributes": {"#status": "status"}, "expression_values": {":status": "In Progress"}}'

# Count number of items in datastore
$ slack datastore count --datastore tasks
```

### Options

```
  -h, --help   help for datastore
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

* [slack](slack)	 - Slack command-line tool
* [slack datastore bulk-delete](slack_datastore_bulk-delete)	 - Delete multiple items from a datastore
* [slack datastore bulk-get](slack_datastore_bulk-get)	 - Get multiple items from a datastore
* [slack datastore bulk-put](slack_datastore_bulk-put)	 - Create or replace a list of items in a datastore
* [slack datastore count](slack_datastore_count)	 - Count the number of items in a datastore
* [slack datastore delete](slack_datastore_delete)	 - Delete an item from a datastore
* [slack datastore get](slack_datastore_get)	 - Get an item from a datastore
* [slack datastore put](slack_datastore_put)	 - Create or replace an item in a datastore
* [slack datastore query](slack_datastore_query)	 - Query a datastore for items
* [slack datastore update](slack_datastore_update)	 - Create or update an item in a datastore

