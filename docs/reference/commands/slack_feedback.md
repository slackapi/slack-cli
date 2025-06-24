# `slack feedback`

Share feedback about your experience or project

## Synopsis

Help us make the Slack Platform better with your feedback

```
slack feedback [flags]
```

## Examples

```
# Choose to give feedback on part of the Slack Platform
$ slack feedback
$ slack feedback --name slack-cli  # Give feedback on the Slack CLI
```

## Options

```
  -h, --help          help for feedback
      --name string   name of the feedback:
                         slack-cli
                         slack-platform
                      
      --no-prompt     run command without prompts
```

## Options inherited from parent commands

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

## SEE ALSO

* [slack](slack)	 - Slack command-line tool

