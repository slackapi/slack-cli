## slack trigger delete

Delete an existing trigger

### Synopsis

Delete an existing trigger

```
slack trigger delete --trigger-id <id> [flags]
```

### Examples

```
# Delete a specific trigger in a selected workspace
$ slack trigger delete --trigger-id Ft01234ABCD

# Delete a specific trigger for an app
$ slack trigger delete --trigger-id Ft01234ABCD --app A0123456
```

### Options

```
  -h, --help                help for delete
      --trigger-id string   the ID of the trigger
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

* [slack trigger](slack_trigger)	 - List details of existing triggers

