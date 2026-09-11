# `slack`

Slack command-line tool

## Description

CLI to create, run, and deploy Slack apps

Get started by reading the docs: [https://docs.slack.dev/tools/slack-cli](https://docs.slack.dev/tools/slack-cli)

```
slack <command> <subcommand> [flags]
```

## Flags

```
      --accessible           use accessible prompts for screen readers
  -a, --app string           use a specific app ID or environment
      --config-dir string    use a custom path for system config directory
  -e, --experiment strings   use the experiment(s) in the command
  -f, --force                ignore warnings and continue executing command
  -h, --help                 help for slack
      --no-color             remove styles and formatting from outputs
  -s, --skip-update          skip checking for latest version of CLI
  -w, --team string          select workspace or organization by team name or ID
      --token string         set the access token associated with a team
  -v, --verbose              print debug logging and additional info
```

## Examples

```
$ slack login   # Log in to your Slack account
$ slack create  # Create a new Slack app
$ slack init    # Initialize an existing Slack app
$ slack run     # Start a local development server
$ slack deploy  # Deploy to the Slack Platform
$ slack docs    # Open Slack developer docs
```

## See also

* [slack activity](/tools/slack-cli/reference/commands/slack_activity/)	 - Display the app activity logs from the Slack Platform
* [slack api](/tools/slack-cli/reference/commands/slack_api/)	 - Call any Slack API method
* [slack app](/tools/slack-cli/reference/commands/slack_app/)	 - Install, uninstall, and list teams with the app installed
* [slack auth](/tools/slack-cli/reference/commands/slack_auth/)	 - Add and remove local team authorizations
* [slack blocks](/tools/slack-cli/reference/commands/slack_blocks/)	 - Build with Block Kit
* [slack collaborator](/tools/slack-cli/reference/commands/slack_collaborator/)	 - Manage app collaborators
* [slack create](/tools/slack-cli/reference/commands/slack_create/)	 - Create a new Slack project
* [slack datastore](/tools/slack-cli/reference/commands/slack_datastore/)	 - Interact with an app's datastore
* [slack delete](/tools/slack-cli/reference/commands/slack_delete/)	 - Delete the app
* [slack deploy](/tools/slack-cli/reference/commands/slack_deploy/)	 - Deploy the app to the Slack Platform
* [slack docs](/tools/slack-cli/reference/commands/slack_docs/)	 - Open Slack developer docs
* [slack doctor](/tools/slack-cli/reference/commands/slack_doctor/)	 - Check and report on system and app information
* [slack env](/tools/slack-cli/reference/commands/slack_env/)	 - Set, unset, or list environment variables
* [slack external-auth](/tools/slack-cli/reference/commands/slack_external-auth/)	 - Adjust settings of external authentication providers
* [slack feedback](/tools/slack-cli/reference/commands/slack_feedback/)	 - Share feedback about your experience or project
* [slack function](/tools/slack-cli/reference/commands/slack_function/)	 - Manage the functions of an app
* [slack init](/tools/slack-cli/reference/commands/slack_init/)	 - Initialize a project to work with the Slack CLI
* [slack install](/tools/slack-cli/reference/commands/slack_install/)	 - Install the app to a team
* [slack list](/tools/slack-cli/reference/commands/slack_list/)	 - List all authorized accounts
* [slack login](/tools/slack-cli/reference/commands/slack_login/)	 - Log in to a Slack account
* [slack logout](/tools/slack-cli/reference/commands/slack_logout/)	 - Log out of a team
* [slack manifest](/tools/slack-cli/reference/commands/slack_manifest/)	 - Print the app manifest of a project or app
* [slack platform](/tools/slack-cli/reference/commands/slack_platform/)	 - Deploy and run apps on the Slack Platform
* [slack project](/tools/slack-cli/reference/commands/slack_project/)	 - Create, manage, and doctor a project
* [slack run](/tools/slack-cli/reference/commands/slack_run/)	 - Start a local server to develop and run the app locally
* [slack samples](/tools/slack-cli/reference/commands/slack_samples/)	 - List available sample apps
* [slack sandbox](/tools/slack-cli/reference/commands/slack_sandbox/)	 - Manage developer sandboxes
* [slack trigger](/tools/slack-cli/reference/commands/slack_trigger/)	 - List details of existing triggers
* [slack uninstall](/tools/slack-cli/reference/commands/slack_uninstall/)	 - Uninstall the app from a team
* [slack upgrade](/tools/slack-cli/reference/commands/slack_upgrade/)	 - Checks for available updates to the CLI or SDK
* [slack version](/tools/slack-cli/reference/commands/slack_version/)	 - Print the version number

