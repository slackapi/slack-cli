## slack env remove

Remove an environment variable from the app

### Synopsis

Remove an environment variable from an app deployed to Slack managed
infrastructure.

If no variable name is provided, you will be prompted to select one.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack env remove <name> [flags]
```

### Examples

```
# Select an environment variable to remove
$ slack env remove
$ slack env remove MAGIC_PASSWORD  # Remove an environment variable
```

### Options

```
  -h, --help          help for remove
      --name string   choose the environment variable name
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

* [slack env](slack_env)	 - Add, remove, or list environment variables

