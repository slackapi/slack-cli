## slackk manifest info

Print the app manifest of a project or app

### Synopsis

Get the manifest of an app using either the "remote" values on app settings
or from the "local" configurations.

The manifest on app settings represents the latest version of the manifest.

Project configurations use the "get-manifest" hook from ".slack/hooks.json".

```
slackk manifest info [flags]
```

### Examples

```
# Print the app manifest from project configurations
$ slackk manifest info

# Print the remote manifest of an app
$ slackk manifest info --app A0123456789

# Print the app manifest gathered from App Config
$ slackk manifest info --source remote
```

### Options

```
  -h, --help            help for info
      --source string   source of the app manifest ("local" or "remote")
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

* [slackk manifest](slackk_manifest.md)	 - Print the app manifest of a project or app

