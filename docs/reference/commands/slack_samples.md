# `slack samples`

List available sample apps

## Description

List and create an app from the available samples

```
slack samples [name] [flags]
```

## Flags

```
  -h, --help              help for samples
      --language string   runtime for the app framework
                            ex: "deno", "node", "python"
      --list              prints samples without interactivity
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
$ slack samples --list --language node  # List Bolt for JavaScript samples
$ slack samples my-project              # Select a sample app to create
```

## See also

* [slack](slack)	 - Slack command-line tool

