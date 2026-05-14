# `slack env list`

List all environment variables of the project

## Description

List the environment variables available to the project.

Commands that run in the context of a project source environment variables from
the ".env" file. This includes the "run" command.

The "deploy" command gathers environment variables from the ".env" file as well
unless the app is using ROSI features.

```
slack env list [flags]
```

## Flags

```
  -h, --help   help for list
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
$ slack env list  # List all environment variables
```

## See also

* [slack env](slack_env)	 - Set, unset, or list environment variables

