# `slack app settings`

Open app settings for configurations

## Description

Open app settings to configure an application in a web browser.

Discovering new features and customizing an app manifest can be done from this
web interface for apps with a "remote" manifest source.

This command does not support apps deployed to Run on Slack infrastructure.

```
slack app settings [flags]
```

## Flags

```
  -h, --help   help for settings
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
# Open app settings for a prompted app
$ slack app settings

# Open app settings for a specific app
$ slack app settings --app A0123456789
```

## See also

* [slack app](slack_app)	 - Install, uninstall, and list teams with the app installed

