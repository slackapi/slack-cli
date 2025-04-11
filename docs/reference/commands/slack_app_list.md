## slack app list

List teams with the app installed

### Synopsis

List all teams that have installed the app

```
slack app list [flags]
```

### Examples

```
$ slack app list  # List all teams with the app installed
```

### Options

```
      --all-org-workspace-grants   display all workspace grants for an app
                                   installed to an organization
  -h, --help                       help for list
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

* [slack app](slack_app)	 - Install, uninstall, and list teams with the app installed

