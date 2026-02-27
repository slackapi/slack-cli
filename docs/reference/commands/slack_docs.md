# `slack docs`

Open Slack developer docs

## Description

Open the Slack developer docs in your browser, with optional search functionality

```
slack docs [flags]
```

## Flags

```
  -h, --help     help for docs
      --search   open Slack docs search page or search with query
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
$ slack docs                       # Open Slack developer docs homepage

# Search Slack developer docs for Block Kit
$ slack docs --search "Block Kit"
$ slack docs --search              # Open Slack docs search page
```

## See also

* [slack](slack)	 - Slack command-line tool

