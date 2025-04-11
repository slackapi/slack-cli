## slack trigger access

Manage who can use your triggers

### Synopsis

Manage who can use your triggers

```
slack trigger access --trigger-id <id> [flags]
```

### Examples

```
# Grant everyone access to run a trigger
$ slack trigger access --trigger-id Ft01234ABCD --everyone

# Grant certain channels access to run a trigger
$ slack trigger access --trigger-id Ft01234ABCD --grant \
    --channels C012345678

# Revoke certain users access to run a trigger
$ slack trigger access --trigger-id Ft01234ABCD --revoke \
    --users USLACKBOT,U012345678
```

### Options

```
  -A, --app-collaborators           grant permission to only app collaborators
  -C, --channels string             a comma-separated list of Slack channel IDs
  -E, --everyone                    grant permission to everyone in your workspace
  -G, --grant                       grant permission to --users or --channels to
                                      run the trigger --trigger-id
  -h, --help                        help for access
      --include-app-collaborators   include app collaborators into named
                                     entities to run the trigger --trigger-id
  -I, --info                        check who has access to the trigger --trigger-id
  -O, --organizations string        a comma-separated list of Slack organization IDs
  -R, --revoke                      revoke permission for --users or --channels to
                                      run the trigger --trigger-id
  -T, --trigger-id string           the ID of the trigger
  -U, --users string                a comma-separated list of Slack user IDs
  -W, --workspaces string           a comma-separated list of Slack workspace IDs
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

