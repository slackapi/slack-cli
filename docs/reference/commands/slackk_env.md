## slackk env

Add, remove, or list environment variables

### Synopsis

Add, remove, or list environment variables for apps deployed to Slack managed
infrastructure.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

Explore more: {{LinkText "https://docs.slack.dev/tools/slack-cli/guides/using-environment-variables-with-the-slack-cli"}}

```
slackk env <subcommand> [flags]
```

### Examples

```
$ slackk env add MAGIC_PASSWORD abracadbra  # Add an environment variable

# List all environment variables
$ slackk env list

# Remove an environment variable
$ slackk env remove MAGIC_PASSWORD
```

### Options

```
  -h, --help   help for env
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

* [slackk](slackk.md)	 - Slack command-line tool
* [slackk env add](slackk_env_add.md)	 - Add an environment variable to the app
* [slackk env list](slackk_env_list.md)	 - List all environment variables for the app
* [slackk env remove](slackk_env_remove.md)	 - Remove an environment variable from the app

