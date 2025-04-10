## slack platform activity

Display the app activity logs from the Slack Platform

### Synopsis

Display the app activity logs from the Slack Platform

```
slack platform activity [flags]
```

### Examples

```
$ slack platform activity     # Display app activity logs for an app
$ slack platform activity -t  # Continuously poll for new activity logs
```

### Options

```
      --component string       component type to filter
      --component-id string    component id to filter
                                 (either a function id or workflow id)
      --event string           event type to filter
  -h, --help                   help for activity
      --idle int               time to poll without results before exiting
                                 in minutes (default 5)
  -i, --interval int           polling interval in seconds (default 3)
      --level string           minimum log level to display (default "info")
                                 (trace, debug, info, warn, error, fatal)
      --limit int              limit the amount of logs retrieved (default 100)
      --max-date-created int   maximum timestamp to filter
                                 (unix timestamp in microseconds)
      --min-date-created int   minimum timestamp to filter
                                 (unix timestamp in microseconds)
      --source string          source (slack or developer) to filter
  -t, --tail                   continuously poll for new activity
      --trace-id string        trace id to filter
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

