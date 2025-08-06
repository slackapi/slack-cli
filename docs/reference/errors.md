# Slack CLI errors reference

Troubleshooting errors can be tricky between your development environment, the
Slack CLI, and those encountered when running your code. Below are some common
ones, as well as a list of the errors the Slack CLI may raise, what they mean,
and some ways to remediate them.

## Slack CLI errors list

### access_denied {#access_denied}

**Message**: You don't have the permission to access the specified resource

**Remediation**: Check with your Slack admin to make sure that you have permission to access the resource.

---

### add_app_to_project_error {#add_app_to_project_error}

**Message**: Couldn't save your app's info to this project

---

### already_logged_out {#already_logged_out}

**Message**: You're already logged out

---

### already_resolved {#already_resolved}

**Message**: The app already has a resolution and cannot be requested

---

### app_add_error {#app_add_error}

**Message**: Couldn't create a new app

---

### app_add_exists {#app_add_exists}

**Message**: App already exists belonging to the team

---

### app_approval_request_denied {#app_approval_request_denied}

**Message**: This app is currently denied for installation

**Remediation**: Reach out to an admin for additional information, or try requesting again with different scopes and outgoing domains

---

### app_approval_request_eligible {#app_approval_request_eligible}

**Message**: This app requires permissions that must be reviewed by an admin before you can install it

---

### app_approval_request_pending {#app_approval_request_pending}

**Message**: This app has requested admin approval to install and is awaiting review

**Remediation**: Reach out to an admin for additional information

---

### app_auth_team_mismatch {#app_auth_team_mismatch}

**Message**: Specified app and team are mismatched

**Remediation**: Try a different combination of --app, --team flags

---

### app_create_error {#app_create_error}

**Message**: Couldn't create your app

---

### app_delete_error {#app_delete_error}

**Message**: Couldn't delete your app

---

### app_deploy_error {#app_deploy_error}

**Message**: Couldn't deploy your app

---

### app_deploy_function_runtime_not_slack {#app_deploy_function_runtime_not_slack}

**Message**: Deployment to Slack is not currently supported for apps with `runOnSlack` set as false

**Remediation**: Learn about building apps with the Deno Slack SDK:

https://docs.slack.dev/tools/deno-slack-sdk

If you are using a Bolt framework, add a deploy hook then run: `slack deploy`

Otherwise start your app for local development with: `slack run`

---

### app_dir_only_fail {#app_dir_only_fail}

**Message**: The app was neither in the app directory nor created on this team/org, and cannot be requested

---

### app_directory_access_error {#app_directory_access_error}

**Message**: Couldn't access app directory

---

### app_flag_required {#app_flag_required}

**Message**: The --app flag must be provided

**Remediation**: Choose a specific app with `--app <app_id>`

---

### app_found {#app_found}

**Message**: An app was found

---

### app_hosted {#app_hosted}

**Message**: App is configured for Run on Slack infrastructure

---

### app_install_error {#app_install_error}

**Message**: Couldn't install your app to a workspace

---

### app_manifest_access_error {#app_manifest_access_error}

**Message**: Couldn't access your app manifest

---

### app_manifest_create_error {#app_manifest_create_error}

**Message**: Couldn't create your app manifest

---

### app_manifest_generate_error {#app_manifest_generate_error}

**Message**: Couldn't generate an app manifest from this project

**Remediation**: Check to make sure you are in a valid Slack project directory and that your project has no compilation errors.

---

### app_manifest_update_error {#app_manifest_update_error}

**Message**: The app manifest was not updated

---

### app_manifest_validate_error {#app_manifest_validate_error}

**Message**: Your app manifest is invalid

---

### app_not_eligible {#app_not_eligible}

**Message**: The specified app is not eligible for this API

---

### app_not_found {#app_not_found}

**Message**: The app was not found

---

### app_not_hosted {#app_not_hosted}

**Message**: App is not configured to be deployed to the Slack platform

**Remediation**: Deploy an app containing workflow automations to Slack managed infrastructure
Read about ROSI: https://docs.slack.dev/workflows/run-on-slack-infrastructure

---

### app_not_installed {#app_not_installed}

**Message**: The provided app must be installed on this team

---

### app_remove_error {#app_remove_error}

**Message**: Couldn't remove your app

---

### app_rename_app {#app_rename_app}

**Message**: Couldn't rename your app

---

### apps_list_error {#apps_list_error}

**Message**: Couldn't get a list of your apps

---

### auth_prod_token_not_found {#auth_prod_token_not_found}

**Message**: Couldn't find a valid auth token for the Slack API

**Remediation**: You need to be logged in to at least 1 production (slack.com) team to use this command. Log into one with the `slack login` command and try again.

---

### auth_timeout_error {#auth_timeout_error}

**Message**: Couldn't receive authorization in the time allowed

**Remediation**: Ensure you have pasted the command in a Slack workspace and accepted the permissions.

---

### auth_token_error {#auth_token_error}

**Message**: Couldn't get a token with an active session

---

### auth_verification_error {#auth_verification_error}

**Message**: Couldn't verify your authorization

---

### bot_invite_required {#bot_invite_required}

**Message**: Your app must be invited to the channel

**Remediation**: Try to find the channel declared the source code of a workflow or function.

Open Slack, join the channel, invite your app, and try the command again.
Learn more: https://slack.com/help/articles/201980108-Add-people-to-a-channel

---

### cannot_abandon_app {#cannot_abandon_app}

**Message**: The last owner cannot be removed

---

### cannot_add_owner {#cannot_add_owner}

**Message**: Unable to add the given user as owner

---

### cannot_count_owners {#cannot_count_owners}

**Message**: Unable to retrieve current app collaborators

---

### cannot_delete_app {#cannot_delete_app}

**Message**: Unable to delete app

---

### cannot_list_collaborators {#cannot_list_collaborators}

**Message**: Calling user is unable to list collaborators

---

### cannot_list_owners {#cannot_list_owners}

**Message**: Calling user is unable to list owners

---

### cannot_remove_collaborators {#cannot_remove_collaborators}

**Message**: User is unable to remove collaborators

---

### cannot_remove_owner {#cannot_remove_owner}

**Message**: Unable to remove the given user

---

### cannot_revoke_org_bot_token {#cannot_revoke_org_bot_token}

**Message**: Revoking org-level bot token is not supported

---

### channel_not_found {#channel_not_found}

**Message**: Couldn't find the specified Slack channel

**Remediation**: Try adding your app as a member to the channel.

---

### cli_autoupdate_error {#cli_autoupdate_error}

**Message**: Couldn't auto-update this command-line tool

**Remediation**: You can manually install the latest version from:
https://docs.slack.dev/tools/slack-cli

---

### cli_config_invalid {#cli_config_invalid}

**Message**: Configuration invalid

**Remediation**: Check your config.json file.

---

### cli_config_location_error {#cli_config_location_error}

**Message**: The .slack/cli.json configuration file is not supported

**Remediation**: This version of the CLI no longer supports this configuration file.
Move the .slack/cli.json file to .slack/hooks.json and try again.

---

### cli_read_error {#cli_read_error}

**Message**: There was an error reading configuration

**Remediation**: Check your config.json file.

---

### cli_update_required {#cli_update_required}

**Message**: Slack API requires the latest version of the Slack CLI

**Remediation**: You can upgrade to the latest version of the Slack CLI using the command: `slack upgrade`

---

### comment_required {#comment_required}

**Message**: Your admin is requesting a reason to approve installation of this app

---

### connected_org_denied {#connected_org_denied}

**Message**: The admin does not allow connected organizations to be named_entities

---

### connected_team_denied {#connected_team_denied}

**Message**: The admin does not allow connected teams to be named_entities

---

### connector_approval_pending {#connector_approval_pending}

**Message**: A connector requires admin approval before it can be installed
Approval is pending review

**Remediation**: Contact your Slack admin about the status of your request

---

### connector_approval_required {#connector_approval_required}

**Message**: A connector requires admin approval before it can be installed

**Remediation**: Request approval for the given connector from your Slack admin

---

### connector_denied {#connector_denied}

**Message**: A connector has been denied for use by an admin

**Remediation**: Contact your Slack admin

---

### connector_not_installed {#connector_not_installed}

**Message**: A connector requires installation before it can be used

**Remediation**: Request installation for the given connector

---

### context_value_not_found {#context_value_not_found}

**Message**: The context value could not be found

---

### credentials_not_found {#credentials_not_found}

**Message**: No authentication found for this team

**Remediation**: Use the command `slack login` to login to this workspace

---

### customizable_input_missing_matching_workflow_input {#customizable_input_missing_matching_workflow_input}

**Message**: Customizable input on the trigger must map to a workflow input of the same name

---

### customizable_input_unsupported_type {#customizable_input_unsupported_type}

**Message**: Customizable input has been mapped to a workflow input of an unsupported type. Only `UserID`, `ChannelId`, and `String` are supported for customizable inputs

---

### customizable_inputs_not_allowed_on_optional_inputs {#customizable_inputs_not_allowed_on_optional_inputs}

**Message**: Customizable trigger inputs must map to required workflow inputs

---

### customizable_inputs_only_allowed_on_link_triggers {#customizable_inputs_only_allowed_on_link_triggers}

**Message**: Customizable inputs are only allowed on link triggers

---

### datastore_error {#datastore_error}

**Message**: An error occurred while accessing your datastore

---

### datastore_missing_primary_key {#datastore_missing_primary_key}

**Message**: The primary key for the datastore is missing

---

### datastore_not_found {#datastore_not_found}

**Message**: The specified datastore could not be found

---

### default_app_access_error {#default_app_access_error}

**Message**: Couldn't access the default app

---

### default_app_setting_error {#default_app_setting_error}

**Message**: Couldn't set this app as the default

---

### deno_not_found {#deno_not_found}

**Message**: Couldn't find the 'deno' language runtime installed on this system

**Remediation**: To install Deno, visit https://deno.land/#installation.

---

### deployed_app_not_supported {#deployed_app_not_supported}

**Message**: A deployed app cannot be used by this command

---

### enterprise_not_found {#enterprise_not_found}

**Message**: The `enterprise` was not found

---

### fail_to_get_teams_for_restricted_user {#fail_to_get_teams_for_restricted_user}

**Message**: Failed to get teams for restricted user

---

### failed_adding_collaborator {#failed_adding_collaborator}

**Message**: Failed writing a collaborator record for this new app

---

### failed_creating_app {#failed_creating_app}

**Message**: Failed to create the app model

---

### failed_datastore_operation {#failed_datastore_operation}

**Message**: Failed while managing datastore infrastructure

**Remediation**: Please try again and reach out to feedback@slack.com if the problem persists.

---

### failed_export {#failed_export}

**Message**: Couldn't export the app manifest

---

### failed_for_some_requests {#failed_for_some_requests}

**Message**: At least one request was not cancelled

---

### failed_to_get_user {#failed_to_get_user}

**Message**: Couldn't find the user to install the app

---

### failed_to_save_extension_logs {#failed_to_save_extension_logs}

**Message**: Couldn't save the logs

---

### feedback_name_invalid {#feedback_name_invalid}

**Message**: The name of the feedback is invalid

**Remediation**: View the feedback options with `slack feedback --help`

---

### feedback_name_required {#feedback_name_required}

**Message**: The name of the feedback is required

**Remediation**: Please provide a `--name <string>` flag or remove the --no-prompt flag
View feedback options with `slack feedback --help`

---

### file_rejected {#file_rejected}

**Message**: Not an acceptable S3 file

---

### forbidden_team {#forbidden_team}

**Message**: The authenticated team cannot use this API

---

### free_team_not_allowed {#free_team_not_allowed}

**Message**: Free workspaces do not support the Slack platform's low-code automation for workflows and functions

**Remediation**: You can install this app if you upgrade your workspace: https://slack.com/pricing.

---

### function_belongs_to_another_app {#function_belongs_to_another_app}

**Message**: The provided function_id does not belong to this app_id

---

### function_not_found {#function_not_found}

**Message**: The specified function couldn't be found

---

### git_clone_error {#git_clone_error}

**Message**: Git failed to clone repository

---

### git_not_found {#git_not_found}

**Message**: Couldn't find Git installed on this system

**Remediation**: To install Git, visit https://github.com/git-guides/install-git.

---

### git_zip_download_error {#git_zip_download_error}

**Message**: Cannot download Git repository as a .zip archive

---

### home_directory_access_failed {#home_directory_access_failed}

**Message**: Failed to read/create .slack/ directory in your home directory

**Remediation**: A Slack directory is required for retrieving/storing auth credentials and config data. Check permissions on your system.

---

### hooks_json_location_error {#hooks_json_location_error}

**Message**: Missing the Slack hooks file from project configurations

**Remediation**: A `.slack/hooks.json` file must be present in the project's `.slack` directory.

---

### hosted_apps_disallow_user_scopes {#hosted_apps_disallow_user_scopes}

**Message**: Hosted apps do not support user scopes

---

### http_request_failed {#http_request_failed}

**Message**: HTTP request failed

---

### http_response_invalid {#http_response_invalid}

**Message**: Received an invalid response from the server

---

### insecure_request {#insecure_request}

**Message**: The method was not called via a `POST` request

---

### installation_denied {#installation_denied}

**Message**: Couldn't install the app because the installation request was denied

**Remediation**: Reach out to one of your App Managers for additional information.

---

### installation_failed {#installation_failed}

**Message**: Couldn't install the app

---

### installation_required {#installation_required}

**Message**: A valid installation of this app is required to take this action

**Remediation**: Install the app with `slack install`

---

### internal_error {#internal_error}

**Message**: An internal error has occurred with the Slack platform

**Remediation**: Please reach out to feedback@slack.com if the problem persists.

---

### invalid_app {#invalid_app}

**Message**: Either the app does not exist or an app created from the provided manifest would not be valid

---

### invalid_app_directory {#invalid_app_directory}

**Message**: This is an invalid Slack app project directory

**Remediation**: A valid Slack project includes the Slack hooks file: .slack/hooks.json

If this is a Slack project, you can initialize it with `slack init`

---

### invalid_app_flag {#invalid_app_flag}

**Message**: The provided --app flag value is not valid

**Remediation**: Specify the environment with --app local or --app deployed
Or choose a specific app with `--app <app_id>`

---

### invalid_app_id {#invalid_app_id}

**Message**: App ID may be invalid for this user account and workspace

**Remediation**: Check to make sure you are signed into the correct workspace for this app and you have the required permissions to perform this action.

---

### invalid_args {#invalid_args}

**Message**: Required arguments either were not provided or contain invalid values

---

### invalid_arguments {#invalid_arguments}

**Message**: Slack API request parameters are invalid

---

### invalid_arguments_customizable_inputs {#invalid_arguments_customizable_inputs}

**Message**: A trigger input parameter with customizable: true cannot be set as hidden or locked, nor have a value provided at trigger creation time

---

### invalid_auth {#invalid_auth}

**Message**: Your user account authorization isn't valid

**Remediation**: Your user account authorization may be expired or does not have permission to access the resource. Try to login to the same user account again using `slack login`.

---

### invalid_challenge {#invalid_challenge}

**Message**: The challenge code is invalid

**Remediation**: The previous slash command and challenge code have now expired. To retry, use `slack login`, paste the slash command in any Slack channel, and enter the challenge code displayed by Slack. It is easiest to copy & paste the challenge code.

---

### invalid_channel_id {#invalid_channel_id}

**Message**: Channel ID specified doesn't exist or you do not have permissions to access it

**Remediation**: Channel ID appears to be formatted correctly. Check if this channel exists on the current team and that you have permissions to access it.

---

### invalid_cursor {#invalid_cursor}

**Message**: Value passed for `cursor` was not valid or is valid no longer

---

### invalid_datastore {#invalid_datastore}

**Message**: Invalid datastore specified in your project

---

### invalid_datastore_expression {#invalid_datastore_expression}

**Message**: The provided expression is not valid

**Remediation**: Verify the expression you provided is valid JSON surrounded by quotations
Use `slack datastore --help` for examples

---

### invalid_distribution_type {#invalid_distribution_type}

**Message**: This function requires distribution_type to be set as named_entities before adding users

---

### invalid_flag {#invalid_flag}

**Message**: The provided flag value is invalid

---

### invalid_interactive_trigger_inputs {#invalid_interactive_trigger_inputs}

**Message**: One or more input parameter types isn't supported by the link trigger type

---

### invalid_manifest {#invalid_manifest}

**Message**: The provided manifest file does not validate against schema. Consult the additional errors field to locate specific issues

---

### invalid_manifest_source {#invalid_manifest_source}

**Message**: A manifest does not exist at the provided source

**Remediation**: Set 'manifest.source' to either "remote" or "local" in .slack/config.json
Read about manifest sourcing with the `slack manifest info --help` command

---

### invalid_parameters {#invalid_parameters}

**Message**: slack_cli_version supplied is invalid

---

### invalid_permission_type {#invalid_permission_type}

**Message**: Permission type must be set to `named_entities` before you can manage users

---

### invalid_refresh_token {#invalid_refresh_token}

**Message**: The given refresh token is invalid

---

### invalid_request_id {#invalid_request_id}

**Message**: The request_id passed is invalid

---

### invalid_resource_id {#invalid_resource_id}

**Message**: The resource_id for the given resource_type is invalid

---

### invalid_resource_type {#invalid_resource_type}

**Message**: The resource_type argument is invalid.

---

### invalid_s3_key {#invalid_s3_key}

**Message**: An internal error occurred

**Remediation**: Please reach out to feedback@slack.com if the problem persists.

---

### invalid_scopes {#invalid_scopes}

**Message**: Some of the provided scopes do not exist

---

### invalid_semver {#invalid_semver}

**Message**: The provided version does not follow semantic versioning

---

### invalid_slack_project_directory {#invalid_slack_project_directory}

**Message**: Current directory is not a Slack project

**Remediation**: Change in to a Slack project directory. A Slack project always includes the Slack hooks file (`.slack/hooks.json`).

---

### invalid_token {#invalid_token}

**Message**: The provided token is not valid

---

### invalid_trigger {#invalid_trigger}

**Message**: Invalid trigger specified in your project

---

### invalid_trigger_access {#invalid_trigger_access}

**Message**: Trigger access can not be configured for more than 10 users

---

### invalid_trigger_config {#invalid_trigger_config}

**Message**: The provided trigger object does not conform to the trigger type's schema

---

### invalid_trigger_event_type {#invalid_trigger_event_type}

**Message**: The provided event type is not allowed

---

### invalid_trigger_inputs {#invalid_trigger_inputs}

**Message**: Required inputs for the referenced function/workflow are not passed

---

### invalid_trigger_type {#invalid_trigger_type}

**Message**: The provided trigger type is not recognized

---

### invalid_user_id {#invalid_user_id}

**Message**: A value passed as a user_id is invalid

---

### invalid_webhook_config {#invalid_webhook_config}

**Message**: Only one of schema or schema_ref should be provided

---

### invalid_webhook_schema_ref {#invalid_webhook_schema_ref}

**Message**: Unable to parse the schema ref

---

### invalid_workflow_app_id {#invalid_workflow_app_id}

**Message**: A value passed as workflow_app_id is invalid or missing

---

### invalid_workflow_id {#invalid_workflow_id}

**Message**: A value passed as a workflow ID is invalid

---

### is_restricted {#is_restricted}

**Message**: Restricted users cannot request

---

### local_app_not_found {#local_app_not_found}

**Message**: Couldn't find the local app

---

### local_app_not_supported {#local_app_not_supported}

**Message**: A local app cannot be used by this command

---

### local_app_removal_error {#local_app_removal_error}

**Message**: Couldn't remove local app

---

### local_app_run_error {#local_app_run_error}

**Message**: Couldn't run app locally

---

### method_not_supported {#method_not_supported}

**Message**: This API method is not supported

---

### mismatched_flags {#mismatched_flags}

**Message**: The provided flags cannot be used together

---

### missing_app_id {#missing_app_id}

**Message**: workflow_app_id is required to update via workflow reference

---

### missing_app_team_id {#missing_app_team_id}

**Message**: team_id is required to create or update this app

---

### missing_challenge {#missing_challenge}

**Message**: Challenge must be supplied

---

### missing_experiment {#missing_experiment}

**Message**: The feature is behind an experiment not toggled on

---

### missing_flag {#missing_flag}

**Message**: An argument must be provided for the flag

---

### missing_function_identifier {#missing_function_identifier}

**Message**: Could not find the given workflow using the specified reference

---

### missing_input {#missing_input}

**Message**: A required value was not supplied as input

---

### missing_options {#missing_options}

**Message**: There are no options to select from

---

### missing_scope {#missing_scope}

**Message**: Your login is out of date

**Remediation**: Run `slack logout` and then `slack login` again.

---

### missing_scopes {#missing_scopes}

**Message**: Additional scopes are required to create this type of trigger

---

### missing_user {#missing_user}

**Message**: The `user` was not found

---

### missing_value {#missing_value}

**Message**: Missing `value` property on an input. You must either provide the value now, or mark this input as `customizable`: `true` and provide the value at the time the trigger is executed.

---

### no_file {#no_file}

**Message**: Couldn't upload your bundled code to server

**Remediation**: Please try again

---

### no_pending_request {#no_pending_request}

**Message**: Pending request not found

---

### no_permission {#no_permission}

**Message**: You are either not a collaborator on this app or you do not have permissions to perform this action

**Remediation**: Contact the app owner to add you as a collaborator

---

### no_token_found {#no_token_found}

**Message**: No tokens found to delete

---

### no_triggers {#no_triggers}

**Message**: There are no triggers installed for this app

---

### no_valid_named_entities {#no_valid_named_entities}

**Message**: None of the provided named entities were valid

---

### not_authed {#not_authed}

**Message**: You are either not logged in or your login session has expired

**Remediation**: Authorize your CLI with `slack login`

---

### not_bearer_token {#not_bearer_token}

**Message**: Incompatible token type provided

---

### not_found {#not_found}

**Message**: Couldn't find row

---

### org_grant_exists {#org_grant_exists}

**Message**: A different org workspace grant already exists for the installed app

---

### org_not_connected {#org_not_connected}

**Message**: One or more of the listed organizations was not connected

---

### org_not_found {#org_not_found}

**Message**: One or more of the listed organizations could not be found

---

### os_not_supported {#os_not_supported}

**Message**: This operating system is not supported

---

### over_resource_limit {#over_resource_limit}

**Message**: Workspace exceeded the maximum number of Run On Slack functions and/or app datastores.

---

### parameter_validation_failed {#parameter_validation_failed}

**Message**: There were problems when validating the inputs against the function parameters. See API response for more details

---

### process_interrupted {#process_interrupted}

**Message**: The process received an interrupt signal

---

### project_compilation_error {#project_compilation_error}

**Message**: An error occurred while compiling your code

---

### project_config_id_not_found {#project_config_id_not_found}

**Message**: The "project_id" property is missing from the project-level configuration file

---

### project_config_manifest_source_error {#project_config_manifest_source_error}

**Message**: Project manifest source is not valid

**Remediation**: Set 'manifest.source' to either "remote" or "local" in .slack/config.json
Read about manifest sourcing with the `slack manifest info --help` command

---

### project_file_update_error {#project_file_update_error}

**Message**: Failed to update project files

---

### prompt_error {#prompt_error}

**Message**: An error occurred while executing prompts

---

### provider_not_found {#provider_not_found}

**Message**: The provided provider_key is invalid

---

### published_app_only {#published_app_only}

**Message**: This action is only permitted for published app IDs

---

### ratelimited {#ratelimited}

**Message**: Too many calls in succession during a short period of time

---

### request_id_or_app_id_is_required {#request_id_or_app_id_is_required}

**Message**: Must include a request_id or app_id

---

### restricted_plan_level {#restricted_plan_level}

**Message**: Your Slack plan does not have access to the requested feature

---

### runtime_not_supported {#runtime_not_supported}

**Message**: The SDK language's executable (deno, node, python, etc) was not found to be installed on the system

---

### sample_create_error {#sample_create_error}

**Message**: Couldn't create app from sample

---

### scopes_exceed_app_config {#scopes_exceed_app_config}

**Message**: Scopes requested exceed app configuration

---

### sdk_config_load_error {#sdk_config_load_error}

**Message**: There was an error while reading the Slack hooks file (`.slack/hooks.json`) or running the `get-hooks` hook

**Remediation**: Run `slack doctor` to check that your system dependencies are up-to-date.

---

### sdk_hook_get_trigger_not_found {#sdk_hook_get_trigger_not_found}

**Message**: The `get-trigger` hook script in `.slack/hooks.json` was not found

**Remediation**: Try defining your trigger by specifying a json file instead.

---

### sdk_hook_invocation_failed {#sdk_hook_invocation_failed}

**Message**: A script hook defined in the Slack Configuration file (`.slack/hooks.json`) returned an error

**Remediation**: Run `slack doctor` to check that your system dependencies are up-to-date.

---

### sdk_hook_not_found {#sdk_hook_not_found}

**Message**: A script in .slack/hooks.json was not found

**Remediation**: Hook scripts are defined in one of these Slack hooks files:
- slack.json
- .slack/hooks.json

Every app requires a Slack hooks file and you can find an example at:
https://github.com/slack-samples/deno-starter-template/blob/main/slack.json

You can create a hooks file manually or with the `slack init` command.

When manually creating the hooks file, you must install the hook dependencies.

---

### service_limits_exceeded {#service_limits_exceeded}

**Message**: Your workspace has exhausted the 10 apps limit for free teams. To create more apps, upgrade your Slack plan at https://my.slack.com/plans

---

### shared_channel_denied {#shared_channel_denied}

**Message**: The team admin does not allow shared channels to be named_entities

---

### slack_auth_error {#slack_auth_error}

**Message**: You are not logged into a team or have not installed an app

**Remediation**: Use the command `slack login` to login and `slack install` to install your app

---

### slack_json_location_error {#slack_json_location_error}

**Message**: The slack.json configuration file is deprecated

**Remediation**: Next major version of the CLI will no longer support this configuration file.
Move the slack.json file to .slack/hooks.json and continue onwards.

---

### slack_slack_json_location_error {#slack_slack_json_location_error}

**Message**: The .slack/slack.json configuration file is deprecated

**Remediation**: Next major version of the CLI will no longer support this configuration file.
Move the .slack/slack.json file to .slack/hooks.json and proceed again.

---

### socket_connection_error {#socket_connection_error}

**Message**: Couldn't connect to Slack over WebSocket

---

### streaming_activity_logs_error {#streaming_activity_logs_error}

**Message**: Failed to stream the most recent activity logs

---

### survey_config_not_found {#survey_config_not_found}

**Message**: Survey config not found

---

### system_config_id_not_found {#system_config_id_not_found}

**Message**: The "system_id" property is missing from the system-level configuration file

---

### system_requirements_failed {#system_requirements_failed}

**Message**: Couldn't verify all system requirements

---

### team_access_not_granted {#team_access_not_granted}

**Message**: There was an issue granting access to the team

---

### team_flag_required {#team_flag_required}

**Message**: The --team flag must be provided

**Remediation**: Choose a specific team with `--team <team_domain>` or `--team <team_id>`

---

### team_list_error {#team_list_error}

**Message**: Couldn't get a list of teams

---

### team_not_connected {#team_not_connected}

**Message**: One or more of the listed teams was not connected by org

---

### team_not_found {#team_not_found}

**Message**: Team could not be found

---

### team_not_on_enterprise {#team_not_on_enterprise}

**Message**: Cannot query team by domain because team is not on an enterprise

---

### team_quota_exceeded {#team_quota_exceeded}

**Message**: Total number of requests exceeded team quota

---

### template_path_not_found {#template_path_not_found}

**Message**: No template app was found at the provided path

---

### token_expired {#token_expired}

**Message**: Your access token has expired

**Remediation**: Use the command `slack login` to authenticate again

---

### token_revoked {#token_revoked}

**Message**: Your token has already been revoked

**Remediation**: Use the command `slack login` to authenticate again

---

### token_rotation_error {#token_rotation_error}

**Message**: An error occurred while rotating your access token

**Remediation**: Use the command `slack login` to authenticate again

---

### too_many_customizable_inputs {#too_many_customizable_inputs}

**Message**: Cannot have more than 10 customizable inputs

---

### too_many_ids_provided {#too_many_ids_provided}

**Message**: Ensure you provide only app_id OR request_id

---

### too_many_named_entities {#too_many_named_entities}

**Message**: Too many named entities passed into the trigger permissions setting

---

### trigger_create_error {#trigger_create_error}

**Message**: Couldn't create a trigger

---

### trigger_delete_error {#trigger_delete_error}

**Message**: Couldn't delete a trigger

---

### trigger_does_not_exist {#trigger_does_not_exist}

**Message**: The trigger provided does not exist

---

### trigger_not_found {#trigger_not_found}

**Message**: The specified trigger cannot be found

---

### trigger_update_error {#trigger_update_error}

**Message**: Couldn't update a trigger

---

### unable_to_delete {#unable_to_delete}

**Message**: There was an error deleting tokens

---

### unable_to_open_file {#unable_to_open_file}

**Message**: Error with file upload

---

### unable_to_parse_json {#unable_to_parse_json}

**Message**: `<json>` Couldn't be parsed as a json object

---

### uninstall_halted {#uninstall_halted}

**Message**: The uninstall process was interrupted

---

### unknown_file_type {#unknown_file_type}

**Message**: Unknown file type, must be application/zip

---

### unknown_function_id {#unknown_function_id}

**Message**: The provided function_id was not found

---

### unknown_method {#unknown_method}

**Message**: The Slack API method does not exist or you do not have permissions to access it

---

### unknown_webhook_schema_ref {#unknown_webhook_schema_ref}

**Message**: Unable to find the corresponding type based on the schema ref

---

### unknown_workflow_id {#unknown_workflow_id}

**Message**: The provided workflow_id was not found for this app

---

### unsupported_file_name {#unsupported_file_name}

**Message**: File name is not supported

---

### untrusted_source {#untrusted_source}

**Message**: Source is by an unknown or untrusted author

**Remediation**: Use --force flag or set trust_unknown_sources: true in config.json file to disable warning

---

### user_already_owner {#user_already_owner}

**Message**: The user is already an owner for this app

---

### user_already_requested {#user_already_requested}

**Message**: The user has a request pending for this app

---

### user_cannot_manage_app {#user_cannot_manage_app}

**Message**: You do not have permissions to install this app

**Remediation**: Reach out to one of your App Managers to request permissions to install apps.

---

### user_id_is_required {#user_id_is_required}

**Message**: Must include a user_id to cancel request for an app with app_id

---

### user_not_found {#user_not_found}

**Message**: User cannot be found

---

### user_removed_from_team {#user_removed_from_team}

**Message**: User removed from team (generated)

---

### workflow_not_found {#workflow_not_found}

**Message**: Workflow not found

---

### yaml_error {#yaml_error}

**Message**: An error occurred while parsing the app manifest YAML file

---

## Additional help

These error codes might reference an error you've encountered, but not provide
enough details for a workaround.

For more help, post to our issue tracker: https://github.com/slackapi/slack-cli/issues
