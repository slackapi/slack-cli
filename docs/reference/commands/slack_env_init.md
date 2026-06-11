# `slack env init`

Initialize environment variables from a template file

## Description

Initialize the project ".env" file by copying from an ".env" template file.

Copies content from either the ".env.sample" or ".env.example" file to the
project ".env" file if those project environment variables don't already exist.

Apps using ROSI features should set environment variables with `slack env set`.

```
slack env init [flags]
```

## Flags

```
  -h, --help   help for init
```

## Global flags

```
      --accessible           use accessible prompts for screen readers
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
$ slack env init  # Initialize environment variables from a template file
```

## See also

* [slack env](slack_env)	 - Set, unset, or list environment variables

