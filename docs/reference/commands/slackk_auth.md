## slackk auth

Add and remove local team authorizations

### Synopsis

Add and remove local team authorizations

```
slackk auth <subcommand> [flags]
```

### Examples

```
$ slackk auth list    # List all authorized accounts
$ slackk auth login   # Log in to a Slack account
$ slackk auth logout  # Log out of a team
```

### Options

```
  -h, --help   help for auth
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

* [slackk](slackk.md)	 - Slack command-line tool
* [slackk auth list](slackk_auth_list.md)	 - List all authorized accounts
* [slackk auth login](slackk_auth_login.md)	 - Log in to a Slack account
* [slackk auth logout](slackk_auth_logout.md)	 - Log out of a team
* [slackk auth revoke](slackk_auth_revoke.md)	 - Revoke an authentication token
* [slackk auth token](slackk_auth_token.md)	 - Collect a service token

