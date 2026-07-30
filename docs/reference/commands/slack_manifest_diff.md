# `slack manifest diff`

Show differences between the project manifest and app settings

## Description

Compare the project manifest with app settings and print any differences.

```
slack manifest diff [flags]
```

## Flags

```
  -h, --help   help for diff
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
# Show differences between project manifest and app settings
$ slack manifest diff

# Show manifest differences without prompts
$ slack manifest diff --app A0123456789 --token xoxp-...
```

## See also

* [slack manifest](slack_manifest)	 - Print the app manifest of a project or app

