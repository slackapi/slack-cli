## slack app link

Add an existing app to the project

### Synopsis

Saves an existing app to a project to be available to other commands.

The provided App ID and Team ID are stored in the apps.json or apps.dev.json
files in the .slack directory of a project.

The environment option decides how an app is handled and where information
should be stored. Production apps should be 'deployed' while apps used for
testing and development should be considered 'local'.

Only one app can exist for each combination of Team ID and environment.

```
slack app link [flags]
```

### Examples

```
# Add an existing app to a project
$ slack app link

# Add a specific app without using prompts
$ slack app link --team T0123456789 --app A0123456789 --environment deployed
```

### Options

```
  -E, --environment string   environment to save existing app (local, deployed)
  -h, --help                 help for link
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

* [slack app](slack_app)	 - Install, uninstall, and list teams with the app installed

