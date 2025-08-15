## slackk function access

Adjust who can access functions published from an app

### Synopsis

Adjust who can {{ToBold "access"}} functions published by an app when building a workflow in
Workflow Builder.

New functions are granted access to {{ToBold "app collaborators"}} by default. This includes
both the {{ToBold "reader"}} and {{ToBold "owner"}} permissions. Access can also be {{ToBold "granted"}} or {{ToBold "revoked"}} to
specific {{ToBold "users"}} or {{ToBold "everyone"}} alongside the {{ToBold "app collaborators"}}.

Workflows that include a function with limited access can still be invoked with
a trigger of the workflow. The {{ToBold "access"}} command applies to Workflow Builder access
only.

```
slackk function access [flags]
```

### Examples

```
# Select a function and choose access options
$ slackk function access

# Share a function with everyone in a workspace
$ slackk function access --name callback_id --everyone

# Revoke function access for multiple users
$ slackk function access --name callback_id --revoke \
    --users USLACKBOT,U012345678,U0RHJTSPQ3

# Lookup access information for a function
$ slackk function access --info
```

### Options

```
  -A, --app-collaborators   grant access to only fellow app collaborators
  -E, --everyone            grant access to everyone in installed workspaces
  -F, --file string         specify access permissions using a file
  -G, --grant               grant access to --users to use --name
  -h, --help                help for access
  -I, --info                check who has access to the function --name
  -N, --name string         the callback_id of a function in your app
  -R, --revoke              revoke access for --users to use --name
  -U, --users string        a comma-separated list of Slack user IDs
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

* [slackk function](slackk_function.md)	 - Manage the functions of an app

