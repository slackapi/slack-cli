# `slack blocks preview`

Preview blocks in the Block Kit Builder

## Description

Open a set of Block Kit blocks in the Block Kit Builder in a web browser.

Blocks can be passed as an argument or piped in through standard input. The
input is a JSON array of blocks or a JSON object with a "blocks" array.

```
slack blocks preview [blocks] [flags]
```

## Flags

```
  -h, --help   help for preview
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
# Preview blocks passed as an argument
$ slack blocks preview '[{"type":"divider"}]'

# Preview blocks piped from a file
$ slack blocks preview < blocks.json
```

## See also

* [slack blocks](slack_blocks)	 - Work with Block Kit blocks

