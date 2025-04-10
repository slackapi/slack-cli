## slack upgrade

Checks for available updates to the CLI or SDK

### Synopsis

Checks for available updates to the CLI or the SDKs of a project

If there are any, then you will be prompted to upgrade

The changelog can be found at [https://docs.slack.dev/changelog/](https://docs.slack.dev/changelog/)

```
slack upgrade [flags]
```

### Examples

```
$ slack upgrade  # Check for any available updates
```

### Options

```
  -h, --help   help for upgrade
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

