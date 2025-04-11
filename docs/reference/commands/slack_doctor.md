## slack doctor

Check and report on system and app information

### Synopsis

Check and report on relevant system (and sometimes app) dependencies

System dependencies can be reviewed from any directory
* This includes operating system information and Deno and Git versions

While app dependencies are only shown within a project directory
* This includes the Deno Slack SDK, API, and hooks versions of an app
* New versions will be listed if there are any updates available

Unfortunately, the doctor command cannot heal all problems

```
slack doctor [flags]
```

### Examples

```
$ slack doctor  # Create a status report of system dependencies
```

### Options

```
  -h, --help   help for doctor
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

