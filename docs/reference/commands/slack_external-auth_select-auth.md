## slack external-auth select-auth

Select developer authentication of a workflow

### Synopsis

Select the saved developer authentication to use when calling external APIs from
functions in a workflow app.

This command is supported for apps deployed to Slack managed infrastructure but
other apps can attempt to run the command with the --force flag.

```
slack external-auth select-auth [flags]
```

### Examples

```
# Select the saved developer authentication in a workflow
$ slack external-auth select-auth --workflow #/workflows/workflow_callback --provider google_provider --external-account user@salesforce.com
```

### Options

```
  -E, --external-account string   external account identifier for the provider
  -h, --help                      help for select-auth
  -p, --provider string           provider of the developer account
  -W, --workflow string           workflow to set developer authentication for
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

