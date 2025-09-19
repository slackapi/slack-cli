# `slack app`

Install, uninstall, and list teams with the app installed

## Description

Install, uninstall, and list teams with the app installed

```
slack app [flags]
```

## Flags

```
  -h, --help   help for app
```

## Global flags

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

## Examples

```
$ slack app install    # Install a production app to a team
$ slack app link       # Link an existing app to the project
$ slack app list       # List all teams with the app installed
$ slack app settings   # Open app settings in a web browser
$ slack app uninstall  # Uninstall an app from a team
$ slack app delete     # Delete an app and app info from a team
```

## See also

* [slack](slack)	 - Slack command-line tool
* [slack app delete](slack_app_delete)	 - Delete the app
* [slack app install](slack_app_install)	 - Install the app to a team
* [slack app link](slack_app_link)	 - Add an existing app to the project
* [slack app list](slack_app_list)	 - List teams with the app installed
* [slack app settings](slack_app_settings)	 - Open app settings for configurations
* [slack app uninstall](slack_app_uninstall)	 - Uninstall the app from a team

