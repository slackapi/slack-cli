## slack auth

Add and remove local team authorizations

### Synopsis

Add and remove local team authorizations

```
slack auth <subcommand> [flags]
```

### Examples

```
$ slack auth list    # List all authorized accounts
$ slack auth login   # Log in to a Slack account
$ slack auth logout  # Log out of a team
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

* [slack](slack)	 - Slack command-line tool
* [slack auth list](slack_auth_list)	 - List all authorized accounts
* [slack auth login](slack_auth_login)	 - Log in to a Slack account
* [slack auth logout](slack_auth_logout)	 - Log out of a team
* [slack auth revoke](slack_auth_revoke)	 - Revoke an authentication token
* [slack auth token](slack_auth_token)	 - Collect a service token

