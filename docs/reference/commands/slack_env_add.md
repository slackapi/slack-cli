## slack env add

Add an environment variable to the app

### Synopsis

Add an environment variable to an app deployed to Slack managed infrastructure.

If a name or value is not provided, you will be prompted to provide these.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack env add <name> <value> [flags]
```

### Examples

```
# Prompt for an environment variable
$ slack env add
$ slack env add MAGIC_PASSWORD abracadbra  # Add an environment variable

# Prompt for an environment variable value
$ slack env add SECRET_PASSWORD
```

### Options

```
  -h, --help           help for add
      --value string   set the environment variable value
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

