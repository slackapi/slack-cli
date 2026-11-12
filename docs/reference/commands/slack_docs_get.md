# `slack docs get`

Print a Slack developer docs page

## Description

Print a page of the Slack developer docs as markdown, return it as JSON, or open
it in your browser.

The page can be a path or a full URL on docs.slack.dev, like a link returned by
the search subcommand.

```
slack docs get <page> [flags]
```

## Flags

```
  -h, --help            help for get
      --output string   output format: text, json, browser (default "text")
```

## Global flags

```
      --accessible           use accessible prompts for screen readers
  -a, --app string           use a specific app ID or environment
      --config-dir string    use a custom path for system config directory
  -e, --experiment strings   use the experiment(s) in the command
  -f, --force                ignore warnings and continue executing command
      --no-color             remove styles and formatting from outputs
  -s, --skip-update          skip checking for latest version of CLI
  -w, --team string          select workspace or organization by team name or ID
      --token string         set the access token associated with a team
                               (use - to read from standard input)
  -v, --verbose              print debug logging and additional info
```

## Examples

```
# Print the Socket Mode guide for Bolt for JavaScript
$ slack docs get /tools/bolt-js/concepts/socket-mode

# Return a page and its links as JSON
$ slack docs get https://docs.slack.dev/reference/methods/chat.postMessage --output=json

# Open a page in your browser
$ slack docs get /block-kit --output=browser
```

## See also

* [slack docs](/tools/slack-cli/reference/commands/slack_docs/)	 - Search and open Slack developer docs

