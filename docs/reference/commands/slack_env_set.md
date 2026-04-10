# `slack env set`

Set an environment variable for the project

## Description

Set an environment variable for the project.

If a name or value is not provided, you will be prompted to provide these.

Commands that run in the context of a project source environment variables from
the ".env" file. This includes the "run" command.

The "deploy" command gathers environment variables from the ".env" file as well
unless the app is using ROSI features.

```
slack env set [name] [value] [flags]
```

## Flags

```
  -h, --help           help for set
      --value string   set the environment variable value
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
# Prompt for an environment variable
$ slack env set
$ slack env set MAGIC_PASSWORD abracadbra  # Set an environment variable

# Prompt for an environment variable value
$ slack env set SECRET_PASSWORD
```

## See also

* [slack env](slack_env)	 - Set, unset, or list environment variables

