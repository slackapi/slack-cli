# `slack project create`

Create a new Slack project

## Description

Create a new Slack project on your local machine from an optional template.

The 'agent' argument is a shortcut to create an AI Agent app. If you want to
name your app 'agent' (not create an AI Agent), use the --name flag instead.

```
slack project create [name | agent <name>] [flags]
```

## Flags

```
  -b, --branch string     name of git branch to checkout
  -h, --help              help for create
      --list              list available app templates
  -n, --name string       name for your app (overrides the name argument)
      --subdir string     subdirectory in the template to use as project
  -t, --template string   template URL for your app
```

## Global flags

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

## Examples

```
# Create a new project from a template
$ slack create my-project

# Create a new AI Agent app
$ slack create agent my-agent-app

# Start a new project from a specific template
$ slack create my-project -t slack-samples/deno-hello-world

# Create a project named 'my-project'
$ slack create --name my-project

# Create from a subdirectory of a template
$ slack create my-project -t org/monorepo --subdir apps/my-app
```

## See also

* [slack project](slack_project)	 - Create, manage, and doctor a project

