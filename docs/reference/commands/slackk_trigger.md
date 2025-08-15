## slackk trigger

List details of existing triggers

### Synopsis

List details of existing triggers

```
slackk trigger [flags]
```

### Examples

```
# Select who can run a trigger
$ slackk trigger access
$ slackk trigger create                           # Create a new trigger

# Delete an existing trigger
$ slackk trigger delete --trigger-id Ft01234ABCD

# Get details for a trigger
$ slackk trigger info --trigger-id Ft01234ABCD

# List details for all existing triggers
$ slackk trigger list

# Update a trigger definition
$ slackk trigger update --trigger-id Ft01234ABCD
```

### Options

```
  -h, --help   help for trigger
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
* [slackk trigger access](slackk_trigger_access.md)	 - Manage who can use your triggers
* [slackk trigger create](slackk_trigger_create.md)	 - Create a trigger for a workflow
* [slackk trigger delete](slackk_trigger_delete.md)	 - Delete an existing trigger
* [slackk trigger info](slackk_trigger_info.md)	 - Get details for a specific trigger
* [slackk trigger list](slackk_trigger_list.md)	 - List details of existing triggers
* [slackk trigger update](slackk_trigger_update.md)	 - Updates an existing trigger

