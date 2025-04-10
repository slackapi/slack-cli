## slack platform run

Start a local server to develop and run the app locally

### Synopsis

Start a local server to develop and run the app locally while watching for file changes

```
slack platform run [flags]
```

### Examples

```
# Start a local development server
$ slack platform run

# Run a local development server with debug activity
$ slack platform run --activity-level debug

# Run a local development server with cleanup
$ slack platform run --cleanup
```

### Options

```
      --activity-level string        activity level to display (default "info")
      --cleanup                      uninstall the local app after exiting
  -h, --help                         help for run
      --hide-triggers                do not list triggers and skip trigger creation prompts
      --no-activity                  hide Slack Platform log activity
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

