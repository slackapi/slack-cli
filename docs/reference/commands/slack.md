## slack

Slack command-line tool

### Synopsis

CLI to create, run, and deploy Slack apps

Get started by reading the docs: [https://tools.slack.dev/deno-slack-sdk](/deno-slack-sdk)

```
slack <command> <subcommand> [flags]
```

### Examples

```
$ slack login   # Log in to your Slack account
$ slack create  # Create a new Slack app
$ slack init    # Initialize an existing Slack app
$ slack run     # Start a local development server
$ slack deploy  # Deploy to the Slack Platform
```

### Options

```
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

### SEE ALSO

* [slack activity](slack_activity)	 - Display the app activity logs from the Slack Platform
* [slack app](slack_app)	 - Install, uninstall, and list teams with the app installed
* [slack auth](slack_auth)	 - Add and remove local team authorizations
* [slack collaborator](slack_collaborator)	 - Manage app collaborators
* [slack create](slack_create)	 - Create a new Slack project
* [slack datastore](slack_datastore)	 - Interact with an app's datastore
* [slack delete](slack_delete)	 - Delete the app
* [slack deploy](slack_deploy)	 - Deploy the app to the Slack Platform
* [slack doctor](slack_doctor)	 - Check and report on system and app information
* [slack env](slack_env)	 - Add, remove, or list environment variables
* [slack external-auth](slack_external-auth)	 - Adjust settings of external authentication providers
* [slack feedback](slack_feedback)	 - Share feedback about your experience or project
* [slack function](slack_function)	 - Manage the functions of an app
* [slack init](slack_init)	 - Initialize a project to work with the Slack CLI
* [slack install](slack_install)	 - Install the app to a team
* [slack list](slack_list)	 - List all authorized accounts
* [slack login](slack_login)	 - Log in to a Slack account
* [slack logout](slack_logout)	 - Log out of a team
* [slack manifest](slack_manifest)	 - Print the app manifest of a project or app
* [slack platform](slack_platform)	 - Deploy and run apps on the Slack Platform
* [slack project](slack_project)	 - Create, manage, and doctor a project
* [slack run](slack_run)	 - Start a local server to develop and run the app locally
* [slack samples](slack_samples)	 - List available sample apps
* [slack trigger](slack_trigger)	 - List details of existing triggers
* [slack uninstall](slack_uninstall)	 - Uninstall the app from a team
* [slack upgrade](slack_upgrade)	 - Checks for available updates to the CLI or SDK
* [slack version](slack_version)	 - Print the version number

