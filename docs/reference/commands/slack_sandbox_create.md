# `slack sandbox create`

Create a developer sandbox

## Description

Create a new Slack developer sandbox

```
slack sandbox create [flags]
```

## Flags

```
      --archive-date string    Explicit archive date in yyyy-mm-dd format. Cannot be used with --archive-ttl
      --archive-ttl string     Time-to-live duration (eg. 1d, 2w, 3mo). Cannot be used with --archive-date
      --domain string          Team domain. Derived from org name if not provided
      --event-code string      Event code for the sandbox
  -h, --help                   help for create
      --locale string          Locale (eg. en-us, languageCode-countryCode)
      --name string            Organization name for the new sandbox
      --owning-org-id string   Enterprise team ID that manages your developer account, if applicable
      --partner                Developers who are part of the Partner program can create partner sandboxes
      --password string        Password used to log into the sandbox
      --template string        Template with sample data to apply to the sandbox (options: default, empty)
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
# Create a sandbox named test-box
$ slack sandbox create --name test-box --password mypass

# Create a temporary sandbox that will be archived in 1 day
$ slack sandbox create --name test-box --password mypass --domain test-box --archive-ttl 1d

# Create a sandbox that will be archived on a specific date
$ slack sandbox create --name test-box --password mypass --domain test-box --archive-date 2025-12-31
```

## See also

* [slack sandbox](slack_sandbox)	 - Manage developer sandboxes

