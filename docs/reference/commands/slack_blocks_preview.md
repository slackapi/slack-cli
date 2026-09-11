# `slack blocks preview`

Preview blocks in Block Kit Builder

## Description

Preview a set of Block Kit blocks with Block Kit Builder in a web browser.

Provide blocks with the --blocks flag.
The input is a JSON array of blocks or a JSON object with a "blocks" array.
Pass - to --blocks, or omit all flags, to read from standard input.

```
slack blocks preview [flags]
```

## Flags

```
      --blocks string   blocks to preview as a JSON array or object
                          (use - to read from standard input)
  -h, --help            help for preview
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
# Preview blocks passed as a flag value
$ slack blocks preview --blocks '[{"type":"divider"}]'

# Preview blocks read from a file
$ slack blocks preview < blocks.json

# Preview blocks read from a redirect and scoped to a team
$ slack blocks preview --team T0123456 --blocks - < blocks.json
```

## See also

* [slack blocks](/tools/slack-cli/reference/commands/slack_blocks/)	 - Build with Block Kit

