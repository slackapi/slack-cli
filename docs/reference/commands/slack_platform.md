## slack platform

Deploy and run apps on the Slack Platform

### Synopsis

Deploy and run apps on the Slack Platform

```
slack platform <subcommand> [flags]
```

### Examples

```
$ slack run                     # Run an app locally in a workspace
$ slack deploy --team T0123456  # Deploy to a specific team
$ slack activity -t             # Continuously poll for new activity logs
```

### Options

```
  -h, --help   help for platform
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
* [slack platform activity](slack_platform_activity)	 - Display the app activity logs from the Slack Platform
* [slack platform deploy](slack_platform_deploy)	 - Deploy the app to the Slack Platform
* [slack platform run](slack_platform_run)	 - Start a local server to develop and run the app locally

