## slack project

Create, manage, and doctor a project

### Synopsis

Create, manage, and doctor a project and its configuration files.

Get started by creating a new project using the **create** command.

Initialize an existing project with CLI support using the **init** command.

Check your project health and diagnose problems with the **doctor** command.

```
slack project <subcommand> [flags]
```

### Examples

```
# Creates a new Slack project from an optional template
$ slack project create

# Initialize an existing project to work with the Slack CLI
$ slack project init

# Creates a new Slack project from the sample gallery
$ slack project samples
```

### Options

```
  -h, --help   help for project
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
* [slack project create](slack_project_create)	 - Create a new Slack project
* [slack project init](slack_project_init)	 - Initialize a project to work with the Slack CLI
* [slack project samples](slack_project_samples)	 - List available sample apps

