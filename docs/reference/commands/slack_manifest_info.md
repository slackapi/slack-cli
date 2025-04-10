## slack manifest info

Print the app manifest of a project or app

### Synopsis

Get the manifest of an app using either the "remote" values on app settings
or from the "local" configurations.

The manifest on app settings represents the latest version of the manifest.

Project configurations use the "get-manifest" hook from ".slack/hooks.json".

```
slack manifest info [flags]
```

### Examples

```
# Print the app manifest from project configurations
$ slack manifest info

# Print the remote manifest of an app
$ slack manifest info --app A0123456789

# Print the app manifest gathered from App Config
$ slack manifest info --source remote
```

### Options

```
  -h, --help            help for info
      --source string   source of the app manifest ("local" or "remote") (default "local")
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

* [slack manifest](slack_manifest)	 - Print the app manifest of a project or app

