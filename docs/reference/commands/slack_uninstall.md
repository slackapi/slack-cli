## slack uninstall

Uninstall the app from a team

### Synopsis

Uninstall the app from a team without deleting the app or its data

```
slack uninstall [flags]
```

### Examples

```
$ slack app uninstall  # Uninstall an app from a team
```

### Options

```
  -h, --help   help for uninstall
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

