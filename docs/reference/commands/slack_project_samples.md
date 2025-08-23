# `slack project samples`

List available sample apps

## Description

List and create an app from the available samples

```
slack project samples [name] [flags]
```

## Flags

```
  -b, --branch string     name of git branch to checkout
  -h, --help              help for samples
      --language string   runtime for the app framework
                            ex: "deno", "node", "python"
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
$ slack samples my-project  # Select a sample app to create
```

## See also

* [slack project](slack_project)	 - Create, manage, and doctor a project

