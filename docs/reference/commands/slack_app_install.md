## slack app install

Install the app to a team

### Synopsis

Install the app to a team

```
slack app install [flags]
```

### Examples

```
$ slack app install                  # Install a production app to a team

# Install a production app to a specific team
$ slack app install --team T0123456
```

### Options

```
  -h, --help                         help for install
      --org-workspace-grant string   grant access to a specific org workspace ID
                                       (or 'all' for all workspaces in the org)
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

