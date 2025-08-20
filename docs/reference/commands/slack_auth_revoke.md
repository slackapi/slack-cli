# `slack auth revoke`

Revoke an authentication token

## Description

Revoke an authentication token

```
slack auth revoke [flags]
```

## Flags

```
  -h, --help   help for revoke
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
$ slack auth revoke --token xoxp-1-4921830...  # Revoke a service token
```

## See also

* [slack auth](slack_auth)	 - Add and remove local team authorizations

