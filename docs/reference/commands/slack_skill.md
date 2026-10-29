# `slack skill`

Print the agent skill for the Slack CLI

## Description

Print the agent skill that teaches coding agents to use the Slack CLI.

The skill is a markdown file with frontmatter that follows the Agent Skills
format. Agents can read the output or save it to a skills directory.

```
slack skill [flags]
```

## Flags

```
  -h, --help   help for skill
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
$ slack skill    # Print the agent skill using the command
$ slack --skill  # Print the agent skill using the flag
```

## See also

* [slack](/tools/slack-cli/reference/commands/slack/)	 - Slack command-line tool

