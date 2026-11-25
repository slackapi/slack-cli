# `slack init`

Initialize a project to work with the Slack CLI

## Description

Initializes a project to support the Slack CLI.

Adds a .slack directory with the following files:
- project-name/.slack
- project-name/.slack/.gitignore
- project-name/.slack/config.json
- project-name/.slack/hooks.json

Adds the Slack CLI hooks dependency to your project:
- Deno:    Unsupported
- Node.js: Updates package.json
- Python:  Updates requirements.txt or pyproject.toml

Installs your project dependencies when supported:
- Deno:    Supported
- Node.js: Supported
- Python:  Supported

Adds an existing app to your project (optional):
- Uses the --app flag to add an existing app from app settings
- Runs the command `slack app link`

```
slack init [flags]
```

## Flags

```
  -E, --environment string   environment to save existing app (local, deployed)
  -h, --help                 help for init
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
                               (use - to read from standard input)
  -v, --verbose              print debug logging and additional info
```

## Examples

```
# Initialize a project
$ slack init

# Initialize a project with an existing app
$ slack init --app A0123456789 --team T0123456789 --environment local
```

## See also

* [slack](/tools/slack-cli/reference/commands/slack/)	 - Slack command-line tool

