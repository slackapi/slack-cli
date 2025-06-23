# `slack external-auth add`

Initiate the OAuth2 flow for a provider

## Synopsis

Initiate the OAuth2 flow for an external auth provider of a workflow app.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack external-auth add [flags]
```

## Examples

```
# Select a provider to initiate the OAuth2 flow for
$ slack external-auth add

# Initiate the OAuth2 flow for the provided provider
$ slack external-auth add -p github
```

## Options

```
  -h, --help              help for add
  -p, --provider string   the external auth Provider Key to add a secret to
```

## Options inherited from parent commands

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

## SEE ALSO

* [slack external-auth](slack_external-auth)	 - Adjust settings of external authentication providers

