# `slack auth logout`

Log out of a team

## Description

Log out of a team, removing any local credentials

```
slack auth logout [flags]
```

## Flags

```
  -A, --all    logout of all workspaces
  -h, --help   help for logout
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
$ slack auth logout        # Select a team to log out of
$ slack auth logout --all  # Log out of all team
```

## See also

* [slack auth](slack_auth)	 - Add and remove local team authorizations

