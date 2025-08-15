## slackk

Slack command-line tool

### Synopsis

{{Emoji "sparkles"}}CLI to create, run, and deploy Slack apps

{{Emoji "books"}}Get started by reading the docs: {{LinkText "https://docs.slack.dev/tools/slack-cli"}}

```
slackk <command> <subcommand> [flags]
```

### Examples

```
$ slackk login   # Log in to your Slack account
$ slackk create  # Create a new Slack app
$ slackk init    # Initialize an existing Slack app
$ slackk run     # Start a local development server
$ slackk deploy  # Deploy to the Slack Platform
```

### Options

```
  -a, --app string           use a specific app ID or environment
      --config-dir string    use a custom path for system config directory
  -e, --experiment strings   use the experiment(s) in the command
  -f, --force                ignore warnings and continue executing command
  -h, --help                 help for slackk
      --no-color             remove styles and formatting from outputs
  -s, --skip-update          skip checking for latest version of CLI
  -w, --team string          select workspace or organization by team name or ID
      --token string         set the access token associated with a team
  -v, --verbose              print debug logging and additional info
```

### SEE ALSO

* [slackk activity](slackk_activity.md)	 - Display the app activity logs from the Slack Platform
* [slackk app](slackk_app.md)	 - Install, uninstall, and list teams with the app installed
* [slackk auth](slackk_auth.md)	 - Add and remove local team authorizations
* [slackk collaborator](slackk_collaborator.md)	 - Manage app collaborators
* [slackk create](slackk_create.md)	 - Create a new Slack project
* [slackk datastore](slackk_datastore.md)	 - Interact with an app's datastore
* [slackk delete](slackk_delete.md)	 - Delete the app
* [slackk deploy](slackk_deploy.md)	 - Deploy the app to the Slack Platform
* [slackk doctor](slackk_doctor.md)	 - Check and report on system and app information
* [slackk env](slackk_env.md)	 - Add, remove, or list environment variables
* [slackk external-auth](slackk_external-auth.md)	 - Adjust settings of external authentication providers
* [slackk feedback](slackk_feedback.md)	 - Share feedback about your experience or project
* [slackk function](slackk_function.md)	 - Manage the functions of an app
* [slackk init](slackk_init.md)	 - Initialize a project to work with the Slack CLI
* [slackk install](slackk_install.md)	 - Install the app to a team
* [slackk list](slackk_list.md)	 - List all authorized accounts
* [slackk login](slackk_login.md)	 - Log in to a Slack account
* [slackk logout](slackk_logout.md)	 - Log out of a team
* [slackk manifest](slackk_manifest.md)	 - Print the app manifest of a project or app
* [slackk platform](slackk_platform.md)	 - Deploy and run apps on the Slack Platform
* [slackk project](slackk_project.md)	 - Create, manage, and doctor a project
* [slackk run](slackk_run.md)	 - Start a local server to develop and run the app locally
* [slackk samples](slackk_samples.md)	 - List available sample apps
* [slackk trigger](slackk_trigger.md)	 - List details of existing triggers
* [slackk uninstall](slackk_uninstall.md)	 - Uninstall the app from a team
* [slackk upgrade](slackk_upgrade.md)	 - Checks for available updates to the CLI or SDK
* [slackk version](slackk_version.md)	 - Print the version number

