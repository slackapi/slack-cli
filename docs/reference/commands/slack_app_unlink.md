# `slack app unlink`

Remove a linked app from the project

## Description

Unlink removes an existing app from the project.

This command removes a saved app ID from the files of a project without deleting
the app from Slack.

```
slack app unlink [flags]
```

## Flags

```
  -h, --help   help for unlink
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
# Remove an existing app from the project
$ slack app unlink

# Remove a specific app without using prompts
$ slack app unlink --app A0123456789
```

## See also

* [slack app](slack_app)	 - Install, uninstall, and list teams with the app installed

