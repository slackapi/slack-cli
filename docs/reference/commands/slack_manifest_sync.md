# `slack manifest sync`

Sync the app manifest between project and app settings

## Description

Compare the local project manifest with app settings, resolve differences, and sync both to the same state.

```
slack manifest sync [flags]
```

## Flags

```
  -h, --help                     help for sync
      --manifest-source string   resolve manifest differences using this source (local or remote)
```

## Global flags

```
      --accessible           use accessible prompts for screen readers
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
# Sync project manifest with app settings
$ slack manifest sync

# Push project manifest to app settings without prompting
$ slack manifest sync --manifest-source=local

# Pull app settings to project manifest without prompting
$ slack manifest sync --manifest-source=remote
```

## See also

* [slack manifest](/tools/slack-cli/reference/commands/slack_manifest/)	 - Print the app manifest of a project or app

