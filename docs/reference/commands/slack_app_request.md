# `slack app request`

Check approval requests to install the app

## Description

Check the status of your most recent request to have the app approved for
install.

Requests are searched on the team of the authenticated account. An account of
a workspace that belongs to an organization also searches that organization,
while an account of an organization searches the organization alone.

Other workspaces of an organization can be searched with the --workspace-ids
flag.

Searches are made with the credentials of an authenticated account chosen
with the --team flag or a prompt.

Apps saved to a project are chosen with a prompt, but any app can be checked
by app ID with the --app flag, which does not require a project.

```
slack app request [flags]
```

## Flags

```
  -h, --help                    help for request
      --workspace-ids strings   also check these workspaces of an organization,
                                with a maximum of 50 workspaces
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
# Check requests to install an app
$ slack app request

# Check requests for an app outside a project
$ slack app request --app A0123456789

# Check requests on certain workspaces of an organization
$ slack app request --workspace-ids T0123456789,T9876543210
```

## See also

* [slack app](/tools/slack-cli/reference/commands/slack_app/)	 - Install, uninstall, and list teams with the app installed

