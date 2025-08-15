## slackk project

Create, manage, and doctor a project

### Synopsis

Create, manage, and doctor a project and its configuration files.

Get started by creating a new project using the {{ToBold "create"}} command.

Initialize an existing project with CLI support using the {{ToBold "init"}} command.

Check your project health and diagnose problems with the {{ToBold "doctor"}} command.

```
slackk project <subcommand> [flags]
```

### Examples

```
# Creates a new Slack project from an optional template
$ slackk project create

# Initialize an existing project to work with the Slack CLI
$ slackk project init

# Creates a new Slack project from the sample gallery
$ slackk project samples
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

* [slackk](slackk.md)	 - Slack command-line tool
* [slackk project create](slackk_project_create.md)	 - Create a new Slack project
* [slackk project init](slackk_project_init.md)	 - Initialize a project to work with the Slack CLI
* [slackk project samples](slackk_project_samples.md)	 - List available sample apps

