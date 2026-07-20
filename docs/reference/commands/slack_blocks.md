# `slack blocks`

Work with Block Kit blocks

## Description

Work with Block Kit blocks, such as previewing them in the Block Kit Builder.

```
slack blocks <subcommand> [flags]
```

## Flags

```
  -h, --help   help for blocks
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
# Preview blocks in the Block Kit Builder
$ slack blocks preview '[{"type":"divider"}]'
```

## See also

* [slack](slack)	 - Slack command-line tool
* [slack blocks preview](slack_blocks_preview)	 - Preview blocks in the Block Kit Builder

