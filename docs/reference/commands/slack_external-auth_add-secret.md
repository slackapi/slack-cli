## slack external-auth add-secret

Add the client secret for a provider

### Synopsis

Add the client secret for an external provider of a workflow app.

This secret will be used when initiating the OAuth2 flow.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack external-auth add-secret [flags]
```

### Examples

```
# Input the client secret for an app and provider
$ slack external-auth add-secret

# Set the client secret for an app and provider
$ slack external-auth add-secret -p github -x ghp_token
```

### Options

```
  -h, --help              help for add-secret
  -p, --provider string   the external auth Provider Key to add a secret to
  -x, --secret string     external auth client secret for the provider
```

### Options inherited from parent commands

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

### SEE ALSO

* [slack external-auth](slack_external-auth)	 - Adjust settings of external authentication providers

