# `slack app install`

Install the app to a team

## Description

Install the app to a team

```
slack app install [flags]
```

## Flags

```
  -E, --environment string           environment of app (local, deployed)
  -h, --help                         help for install
      --org-workspace-grant string   grant access to a specific org workspace ID
                                       (or 'all' for all workspaces in the org)
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
# Install a production app to a team
$ slack app install

# Install a production app to a specific team
$ slack app install --team T0123456 --environment deployed

# Install a local dev app to a specific team
$ slack app install --team T0123456 --environment local
```

## See also

* [slack app](slack_app)	 - Install, uninstall, and list teams with the app installed

