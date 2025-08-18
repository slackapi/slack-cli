# `slack collaborator add`

Add a new collaborator to the app

## Description

Add a collaborator to your app by Slack email address or user ID

```
slack collaborator add [email|user_id] [flags]
```

## Flags

```
  -h, --help   help for add
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
$ slack collaborator add                # Add a collaborator via prompt
$ slack collaborator add bot@slack.com  # Add a collaborator from email
$ slack collaborator add USLACKBOT      # Add a collaborator by user ID
```

## See also

* [slack collaborator](slack_collaborator)	 - Manage app collaborators

