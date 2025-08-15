## slackk trigger list

List details of existing triggers

### Synopsis

List details of existing triggers

```
slackk trigger list [flags]
```

### Examples

```
# List details for all existing triggers
$ slackk trigger list

# List triggers for a specific app
$ slackk trigger list --team T0123456 --app local
```

### Options

```
  -h, --help          help for list
  -L, --limit int     Limit the number of triggers to show (default 4)
  -T, --type string   Only display triggers of the given type, can be one of 'all', 'shortcut', 'event', 'scheduled', 'webhook', and 'external' (default "all")
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

* [slackk trigger](slackk_trigger.md)	 - List details of existing triggers

