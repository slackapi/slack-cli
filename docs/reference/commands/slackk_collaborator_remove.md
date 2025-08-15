## slackk collaborator remove

Remove a collaborator from an app

### Synopsis

Remove a collaborator from an app by Slack email address or user ID

```
slackk collaborator remove [email|user_id] [flags]
```

### Examples

```
# Remove collaborator on prompt
$ slackk collaborator remove
$ slackk collaborator remove bot@slack.com  # Remove collaborator by email
$ slackk collaborator remove USLACKBOT      # Remove collaborator using ID
```

### Options

```
  -h, --help   help for remove
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

* [slackk collaborator](slackk_collaborator.md)	 - Manage app collaborators

