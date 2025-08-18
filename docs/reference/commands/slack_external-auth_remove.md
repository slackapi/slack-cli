# `slack external-auth remove`

Remove the saved tokens for a provider

## Description

Remove tokens saved to external authentication providers of a workflow app.

Existing tokens are only removed from your app, but are not revoked or deleted!
Tokens must be invalidated using the provider's developer console or via APIs.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack external-auth remove [flags]
```

## Flags

```
  -A, --all               remove tokens for all providers or the specified provider
  -h, --help              help for remove
  -p, --provider string   the external auth Provider Key to remove a token for
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
# Remove a token from the selected provider
$ slack external-auth remove

# Remove a token from the specified provider
$ slack external-auth remove -p github

# Remove all tokens from the specified provider
$ slack external-auth remove --all -p github

# Remove all tokens from all providers
$ slack external-auth remove --all
```

## See also

* [slack external-auth](slack_external-auth)	 - Adjust settings of external authentication providers

