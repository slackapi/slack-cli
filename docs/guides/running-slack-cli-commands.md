---
sidebar_label: Running commands
slug: /tools/slack-cli/guides/running-slack-cli-commands
---

# Running Slack CLI commands

The Slack CLI allows you to interact with your apps via the command line. Using the main command `slack`, you can create, run, and deploy apps, as well as create triggers and query datastores.

:::info[Running `slack help` will display available commands in your terminal window.]

:::

Use commands as follows (unless otherwise noted):

```
slack <command> <subcommand> [flags]
```

To view global flags and each subcommands flags, run the following in your terminal:

```
slack <subcommand> --help
```

## Commands overview {#overview}

Below you'll find all the commands and subcommands for the Slack CLI. Each one has its own reference page.

| Command |  Description |
| :--- | :--- |
| [`slack activity`](/tools/slack-cli/reference/commands/slack_activity) |  Display the app activity logs from the Slack Platform
| [`slack app`](/tools/slack-cli/reference/commands/slack_app) |  Install, uninstall, and list teams with the app installed
| [`slack auth`](/tools/slack-cli/reference/commands/slack_auth) |  Add and remove local team authorizations
| [`slack collaborator`](/tools/slack-cli/reference/commands/slack_collaborator) |  Manage app collaborators
| [`slack create`](/tools/slack-cli/reference/commands/slack_create) |  Create a Slack project
| [`slack datastore`](/tools/slack-cli/reference/commands/slack_datastore) |  Query an app's datastore
| [`slack delete`](/tools/slack-cli/reference/commands/slack_delete) |  Delete the app
| [`slack deploy`](/tools/slack-cli/reference/commands/slack_deploy) |  Deploy the app to the Slack Platform
| [`slack doctor`](/tools/slack-cli/reference/commands/slack_doctor) |  Check and report on system and app information
| [`slack env`](/tools/slack-cli/reference/commands/slack_env) |  Add, remove, and list environment variables
| [`slack external-auth`](/tools/slack-cli/reference/commands/slack_external-auth) |  Add and remove external authorizations and client secrets for providers in your app
| [`slack feedback`](/tools/slack-cli/reference/commands/slack_feedback) |  Share feedback about your experience or project
| [`slack function`](/tools/slack-cli/reference/commands/slack_function) |  Manage the functions of an app
| [`slack install`](/tools/slack-cli/reference/commands/slack_install) |  Install the app to a team
| [`slack list`](/tools/slack-cli/reference/commands/slack_list) |  List all authorized accounts
| [`slack login`](/tools/slack-cli/reference/commands/slack_login) |  Log in to a Slack account
| [`slack logout`](/tools/slack-cli/reference/commands/slack_logout) |  Log out of a team
| [`slack manifest`](/tools/slack-cli/reference/commands/slack_manifest) |  Print the app manifest of a project or app
| [`slack platform`](/tools/slack-cli/reference/commands/slack_platform) |  Deploy and run apps on the Slack Platform
| [`slack run`](/tools/slack-cli/reference/commands/slack_run) |  Start a local server to develop and run the app locally
| [`slack samples`](/tools/slack-cli/reference/commands/slack_samples) |  List available sample apps
| [`slack trigger`](/tools/slack-cli/reference/commands/slack_trigger) |  List details of existing triggers
| [`slack uninstall`](/tools/slack-cli/reference/commands/slack_uninstall) |  Uninstall the app from a team
| [`slack upgrade`](/tools/slack-cli/reference/commands/slack_upgrade) |  Checks for available updates to the CLI or SDK
| [`slack version`](/tools/slack-cli/reference/commands/slack_version) |  Print the version number