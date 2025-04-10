## slack collaborator

Manage app collaborators

### Synopsis

Manage app collaborators

```
slack collaborator <subcommand> [flags]
```

### Examples

```
$ slack collaborator add bots@slack.com  # Add a collaborator from email
$ slack collaborator list                # List all of the collaborators

# Remove a collaborator by user ID
$ slack collaborator remove USLACKBOT
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

* [slack](slack)	 - Slack command-line tool
* [slack collaborator add](slack_collaborator_add)	 - Add a new collaborator to the app
* [slack collaborator list](slack_collaborator_list)	 - List all collaborators of an app
* [slack collaborator remove](slack_collaborator_remove)	 - Remove a collaborator from an app

