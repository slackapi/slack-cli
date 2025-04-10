## slack app

Install, uninstall, and list teams with the app installed

### Synopsis

Install, uninstall, and list teams with the app installed

```
slack app [flags]
```

### Examples

```
$ slack install    # Install a production app to a team
$ slack link       # Link an existing app to the project
$ slack list       # List all teams with the app installed
$ slack uninstall  # Uninstall an app from a team
$ slack delete     # Delete an app and app info from a team
```

### Options

```
  -h, --help   help for app
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
* [slack app delete](slack_app_delete)	 - Delete the app
* [slack app install](slack_app_install)	 - Install the app to a team
* [slack app link](slack_app_link)	 - Add an existing app to the project
* [slack app list](slack_app_list)	 - List teams with the app installed
* [slack app uninstall](slack_app_uninstall)	 - Uninstall the app from a team

