## slack init

Initialize a project to work with the Slack CLI

### Synopsis

Initializes a project to support the Slack CLI.

Adds a .slack directory with the following files:
- project-name/.slack
- project-name/.slack/.gitignore
- project-name/.slack/config.json
- project-name/.slack/hooks.json

Adds the Slack CLI hooks dependency to your project:
- Deno:    Unsupported
- Node.js: Updates package.json
- Python:  Updates requirements.txt

Installs your project dependencies when supported:
- Deno:    Supported
- Node.js: Supported
- Python:  Unsupported

Adds an existing app to your project (optional):
- Prompts to add an existing app from app settings
- Runs the command `slack app link`

```
slack init [flags]
```

### Examples

```
$ slack init  # Initialize a project
```

### Options

```
  -h, --help   help for init
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

