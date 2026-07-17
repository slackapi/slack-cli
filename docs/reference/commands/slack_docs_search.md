# `slack docs search`

Search Slack developer docs

## Description

Search the Slack developer docs and return results in text, JSON, or browser
format.

Results can be filtered to a single category with the --category flag. Available
categories: guides, reference, changelog, python, javascript, java, slack_cli, slack_github_action, deno_slack_sdk.

```
slack docs search [query] [flags]
```

## Flags

```
      --category string   filter results by category: guides, reference, changelog, python, javascript, java, slack_cli, slack_github_action, deno_slack_sdk
  -h, --help              help for search
      --limit int         maximum number of text or json search results to return (default 20)
      --output string     output format: text, json, browser (default "text")
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
  -v, --verbose              print debug logging and additional info
```

## Examples

```
# Search docs and return text results
$ slack docs search "Block Kit"

# Search docs and open results in browser
$ slack docs search "webhooks" --output=browser

# Search docs with limited JSON results
$ slack docs search "api" --output=json --limit=5

# Search only the API reference docs
$ slack docs search "chat.postMessage" --category=reference
```

## See also

* [slack docs](slack_docs)	 - Open Slack developer docs

