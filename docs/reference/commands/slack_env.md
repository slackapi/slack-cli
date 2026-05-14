# `slack env`

Set, unset, or list environment variables

## Description

Set, unset, or list environment variables for the project.

Commands that run in the context of a project source environment variables from
the ".env" file. This includes the "run" command.

The "deploy" command gathers environment variables from the ".env" file as well
unless the app is using ROSI features.

Explore more: [https://docs.slack.dev/tools/slack-cli/guides/using-environment-variables-with-the-slack-cli](https://docs.slack.dev/tools/slack-cli/guides/using-environment-variables-with-the-slack-cli)

```
slack env <subcommand> [flags]
```

## Flags

```
  -h, --help   help for env
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
$ slack env set MAGIC_PASSWORD abracadbra  # Set an environment variable

# List all environment variables
$ slack env list
$ slack env unset MAGIC_PASSWORD           # Unset an environment variable
```

## See also

* [slack](slack)	 - Slack command-line tool
* [slack env list](slack_env_list)	 - List all environment variables of the project
* [slack env set](slack_env_set)	 - Set an environment variable for the project
* [slack env unset](slack_env_unset)	 - Unset an environment variable from the project

