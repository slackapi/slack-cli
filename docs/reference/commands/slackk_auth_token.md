## slackk auth token

Collect a service token

### Synopsis

Log in to a Slack account in your team

```
slackk auth token [flags]
```

### Examples

```
# Create a service token with prompts
$ slackk auth token

# Gather a service token without prompts, this returns a ticket
$ slackk auth token --no-prompt

# Complete authentication using a ticket and challenge code
$ slackk auth token --challenge 6d0a31c9 --ticket ISQWLiZT0OtMLO3YWNTJO0...
```

### Options

```
      --challenge string   provide a challenge code for pre-authenticated login
  -h, --help               help for token
      --no-prompt          login without prompts using ticket and challenge code
      --ticket string      provide an auth ticket value
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

* [slackk auth](slackk_auth.md)	 - Add and remove local team authorizations

