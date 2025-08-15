## slackk collaborator

Manage app collaborators

### Synopsis

Manage app collaborators

```
slackk collaborator <subcommand> [flags]
```

### Examples

```
$ slackk collaborator add bots@slack.com  # Add a collaborator from email
$ slackk collaborator list                # List all of the collaborators

# Remove a collaborator by user ID
$ slackk collaborator remove USLACKBOT
```

### Options

```
  -h, --help   help for collaborator
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
* [slackk collaborator add](slackk_collaborator_add.md)	 - Add a new collaborator to the app
* [slackk collaborator list](slackk_collaborator_list.md)	 - List all collaborators of an app
* [slackk collaborator remove](slackk_collaborator_remove.md)	 - Remove a collaborator from an app

