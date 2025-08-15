## slackk app

Install, uninstall, and list teams with the app installed

### Synopsis

Install, uninstall, and list teams with the app installed

```
slackk app [flags]
```

### Examples

```
$ slackk app install    # Install a production app to a team
$ slackk app link       # Link an existing app to the project
$ slackk app list       # List all teams with the app installed
$ slackk app settings   # Open app settings in a web browser
$ slackk app uninstall  # Uninstall an app from a team
$ slackk app delete     # Delete an app and app info from a team
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

* [slackk](slackk.md)	 - Slack command-line tool
* [slackk app delete](slackk_app_delete.md)	 - Delete the app
* [slackk app install](slackk_app_install.md)	 - Install the app to a team
* [slackk app link](slackk_app_link.md)	 - Add an existing app to the project
* [slackk app list](slackk_app_list.md)	 - List teams with the app installed
* [slackk app settings](slackk_app_settings.md)	 - Open app settings for configurations
* [slackk app uninstall](slackk_app_uninstall.md)	 - Uninstall the app from a team

