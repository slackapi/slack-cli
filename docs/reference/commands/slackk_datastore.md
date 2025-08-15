## slackk datastore

Interact with an app's datastore

### Synopsis

Interact with the items stored in an app's datastore.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

Discover the datastores: {{LinkText "https://docs.slack.dev/tools/deno-slack-sdk/guides/using-datastores"}}

```
slackk datastore <subcommand> <expression> [flags]
```

### Examples

```
# Add a new entry to the datastore
$ slackk datastore put --datastore tasks '{"item": {"id": "42", "description": "Create a PR", "status": "Done"}}'

# Add two new entries to the datastore
$ slackk datastore bulk-put --datastore tasks '{"items": [{"id": "12", "description": "Create a PR", "status": "Done"}, {"id": "42", "description": "Approve a PR", "status": "Pending"}]}'

# Update the entry in the datastore
$ slackk datastore update --datastore tasks '{"item": {"id": "42", "description": "Create a PR", "status": "Done"}}'

# Get an item from the datastore
$ slackk datastore get --datastore tasks '{"id": "42"}'

# Get two items from datastore
$ slackk datastore bulk-get --datastore tasks '{"ids": ["12", "42"]}'

# Remove an item from the datastore
$ slackk datastore delete --datastore tasks '{"id": "42"}'

# Remove two items from the datastore
$ slackk datastore bulk-delete --datastore tasks '{"ids": ["12", "42"]}'

# Query the datastore for specific items
$ slackk datastore query --datastore tasks '{"expression": "#status = :status", "expression_attributes": {"#status": "status"}, "expression_values": {":status": "In Progress"}}'

# Count number of items in datastore
$ slackk datastore count --datastore tasks
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

* [slackk](slackk.md)	 - Slack command-line tool
* [slackk datastore bulk-delete](slackk_datastore_bulk-delete.md)	 - Delete multiple items from a datastore
* [slackk datastore bulk-get](slackk_datastore_bulk-get.md)	 - Get multiple items from a datastore
* [slackk datastore bulk-put](slackk_datastore_bulk-put.md)	 - Create or replace a list of items in a datastore
* [slackk datastore count](slackk_datastore_count.md)	 - Count the number of items in a datastore
* [slackk datastore delete](slackk_datastore_delete.md)	 - Delete an item from a datastore
* [slackk datastore get](slackk_datastore_get.md)	 - Get an item from a datastore
* [slackk datastore put](slackk_datastore_put.md)	 - Create or replace an item in a datastore
* [slackk datastore query](slackk_datastore_query.md)	 - Query a datastore for items
* [slackk datastore update](slackk_datastore_update.md)	 - Create or update an item in a datastore

