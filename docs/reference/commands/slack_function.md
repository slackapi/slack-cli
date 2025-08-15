# `slack function`

Manage the functions of an app

## Description

Functions are pieces of logic that complete the puzzle of workflows in Workflow
Builder. Whatever that puzzle might be.

Inspect and configure the custom functions included in an app with this command.
Functions can be added as a step in Workflow Builder and shared among teammates.

Learn more about functions: [https://docs.slack.dev/tools/deno-slack-sdk/guides/creating-functions](https://docs.slack.dev/tools/deno-slack-sdk/guides/creating-functions)

```
slack function <subcommand> [flags]
```

## Flags

```
  -h, --help   help for function
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
# Select a function and choose distribution options
$ slack function distribute

# Distribute a function to everyone in a workspace
$ slack function distribute --name callback_id --everyone

# Lookup the distribution information for a function
$ slack function distribute --info
```

## See also

* [slack](slack)	 - Slack command-line tool
* [slack function access](slack_function_access)	 - Adjust who can access functions published from an app

