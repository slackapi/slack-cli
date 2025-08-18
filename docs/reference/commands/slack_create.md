# `slack create`

Create a new Slack project

## Description

Create a new Slack project on your local machine from an optional template

```
slack create [name] [flags]
```

## Flags

```
  -b, --branch string     name of git branch to checkout
  -h, --help              help for create
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

# Start a new project from a specific template
$ slack create my-project -t slack-samples/deno-hello-world
```

## See also

* [slack](slack)	 - Slack command-line tool

