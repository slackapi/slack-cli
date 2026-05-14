# `slack api`

Call any Slack API method

## Description

Call any Slack API method directly.

The method argument is the Slack API method name (e.g., "chat.postMessage").
Parameters are passed as key=value pairs, a JSON body, or via flags.

Body format is auto-detected from positional arguments:
  - Multiple key=value args: form-encoded (token in request body)
  - Single arg starting with { or [: JSON (Bearer token in header)
  - No args: token sent in Authorization header

Use --json to explicitly send a JSON body, or --data for a form-encoded body string.

Token resolution (in priority order):
  1. --token flag              Explicit token value
  2. --app flag                Install app and use bot token (in project)
  3. SLACK_BOT_TOKEN env var   Bot token (set during slack deploy)
  4. SLACK_USER_TOKEN env var  User token
  5. App prompt (in project)   Select installed app and use bot token

See all methods at: https://docs.slack.dev/reference/methods

```
slack api <method> [key=value ...] [flags]
```

## Flags

```
      --data string      form-encoded request body string (e.g. "key1=val1&key2=val2")
  -H, --header strings   additional HTTP headers (format: "Key: Value")
  -h, --help             help for api
  -i, --include          include HTTP status code and response headers in output
      --json string      JSON request body (uses Bearer token in Authorization header)
  -X, --method string    HTTP method for the request (default "POST")
```

## Global flags

```
      --accessible           use accessible prompts for screen readers
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
# Test your API connection
$ slack api api.test

# Check authentication
$ slack api auth.test

# Add a bookmark to a channel
$ slack api bookmarks.add channel_id=C0123456 title=Docs link=https://example.com

# Send a message to a channel using form-encoded string
$ slack api chat.postMessage channel=C0123456 text="Hello"

# Send a message to a channel using JSON
$ slack api chat.postMessage --json '{"channel":"C0123456","text":"Hello"}'

# Update a message
$ slack api chat.update channel=C0123456 ts=1234567890.123456 text="Updated"

# Create a channel
$ slack api conversations.create name=new-channel

# Fetch messages from a channel
$ slack api conversations.history channel=C0123456

# Get channel details
$ slack api conversations.info channel=C0123456

# List channels
$ slack api conversations.list

# List members in a channel
$ slack api conversations.members channel=C0123456

# Upload a file
$ slack api files.upload channels=C0123456 filename=report.csv

# Pin a message
$ slack api pins.add channel=C0123456 timestamp=1234567890.123456

# Add an emoji reaction
$ slack api reactions.add channel=C0123456 timestamp=1234567890.123456 name=thumbsup

# List reactions for a user
$ slack api reactions.list user=U0123456

# Get user details
$ slack api users.info user=U0123456

# List workspace members
$ slack api users.list

# Get a user's profile
$ slack api users.profile.get user=U0123456

# Open a modal view
$ slack api views.open trigger_id=T0123456 view={...}

# Update a modal view
$ slack api views.update view_id=V0123456 view={...}
```

## See also

* [slack](slack)	 - Slack command-line tool

