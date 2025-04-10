## slack external-auth

Adjust settings of external authentication providers

### Synopsis

Adjust external authorization and authentication providers of a workflow app.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

Explore providers: [Integrating with external auth providers](/guides/integrating-with-services-requiring-external-authentication/)

```
slack external-auth <subcommand> [flags]
```

### Examples

```
# Initiate OAuth2 flow for a selected provider
$ slack external-auth add

# Set client secret for an app and provider
$ slack external-auth add-secret

# Remove authorization for a specific provider
$ slack external-auth remove

# Select authorization for a specific provider in a workflow
$ slack external-auth select-auth
```

### Options

```
  -h, --help   help for external-auth
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

* [slack](slack)	 - Slack command-line tool
* [slack external-auth add](slack_external-auth_add)	 - Initiate the OAuth2 flow for a provider
* [slack external-auth add-secret](slack_external-auth_add-secret)	 - Add the client secret for a provider
* [slack external-auth remove](slack_external-auth_remove)	 - Remove the saved tokens for a provider
* [slack external-auth select-auth](slack_external-auth_select-auth)	 - Select developer authentication of a workflow

