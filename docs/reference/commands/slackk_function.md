## slackk function

Manage the functions of an app

### Synopsis

Functions are pieces of logic that complete the puzzle of workflows in Workflow
Builder. Whatever that puzzle might be.

Inspect and configure the custom functions included in an app with this command.
Functions can be added as a step in Workflow Builder and shared among teammates.

Learn more about functions: {{LinkText "https://docs.slack.dev/tools/deno-slack-sdk/guides/creating-functions"}}

```
slackk function <subcommand> [flags]
```

### Examples

```
# Select a function and choose distribution options
$ slackk function distribute

# Distribute a function to everyone in a workspace
$ slackk function distribute --name callback_id --everyone

# Lookup the distribution information for a function
$ slackk function distribute --info
```

### Options

```
  -h, --help   help for function
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
* [slackk function access](slackk_function_access.md)	 - Adjust who can access functions published from an app

