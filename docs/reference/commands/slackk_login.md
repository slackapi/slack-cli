## slackk login

Log in to a Slack account

### Synopsis

Log in to a Slack account in your team

```
slackk login [flags]
```

### Examples

```
# Login to a Slack account with prompts
$ slackk auth login

# Login to a Slack account without prompts, this returns a ticket
$ slackk auth login --no-prompt

# Complete login using ticket and challenge code
$ slackk auth login --challenge 6d0a31c9 --ticket ISQWLiZT0OtMLO3YWNTJO0...

# Login with a user token
$ slackk auth login --token xoxp-...
```

### Options

```
      --challenge string   provide a challenge code for pre-authenticated login
  -h, --help               help for login
      --no-prompt          login without prompts using ticket and challenge code
      --ticket string      provide an auth ticket value
      --token string       provide a token for a pre-authenticated login
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
  -v, --verbose              print debug logging and additional info
```

### SEE ALSO

* [slackk](slackk.md)	 - Slack command-line tool

