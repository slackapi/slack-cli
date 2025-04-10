## slack platform deploy

Deploy the app to the Slack Platform

### Synopsis

Deploy the app to the Slack Platform

```
slack platform deploy [flags]
```

### Examples

```
# Select the workspace to deploy to
$ slack platform deploy
$ slack platform deploy --team T0123456  # Deploy to a specific team
```

### Options

```
  -h, --help                         help for deploy
      --hide-triggers                do not list triggers and skip trigger creation prompts
      --org-workspace-grant string   grant access to a specific org workspace ID
                                       (or 'all' for all workspaces in the org)
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

* [slack platform](slack_platform)	 - Deploy and run apps on the Slack Platform

