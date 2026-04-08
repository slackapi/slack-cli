# `slack docs`

Open Slack developer docs

## Description

Open the Slack developer docs in your browser or search them using the search subcommand

```
slack docs [flags]
```

## Flags

```
  -h, --help   help for docs
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
# Open Slack developer docs homepage
$ slack docs

# Search Slack developer docs for Block Kit
$ slack docs search "Block Kit"

# Search docs and open results in browser
$ slack docs search "Block Kit" --output=browser
```

## See also

* [slack](slack)	 - Slack command-line tool
* [slack docs search](slack_docs_search)	 - Search Slack developer docs

