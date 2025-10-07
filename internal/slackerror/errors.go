// Copyright 2022-2025 Salesforce, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package slackerror

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/style"
)

const (
	ErrAccessDenied                                  = "access_denied"
	ErrAddAppToProject                               = "add_app_to_project_error"
	ErrAlreadyLoggedOut                              = "already_logged_out"
	ErrAlreadyResolved                               = "already_resolved"
	ErrAppsList                                      = "apps_list_error"
	ErrAppAdd                                        = "app_add_error"
	ErrAppApprovalRequestDenied                      = "app_approval_request_denied"
	ErrAppApprovalRequestEligible                    = "app_approval_request_eligible"
	ErrAppApprovalRequestPending                     = "app_approval_request_pending"
	ErrAppAuthTeamMismatch                           = "app_auth_team_mismatch"
	ErrAppCreate                                     = "app_create_error"
	ErrAppDelete                                     = "app_delete_error"
	ErrAppDeploy                                     = "app_deploy_error"
	ErrAppDeployNotSlackHosted                       = "app_deploy_function_runtime_not_slack"
	ErrAppDirectoryAccess                            = "app_directory_access_error"
	ErrAppDirOnlyFail                                = "app_dir_only_fail"
	ErrAppExists                                     = "app_add_exists"
	ErrAppFlagRequired                               = "app_flag_required"
	ErrAppFound                                      = "app_found"
	ErrAppHosted                                     = "app_hosted"
	ErrAppInstall                                    = "app_install_error"
	ErrAppManifestAccess                             = "app_manifest_access_error"
	ErrAppManifestCreate                             = "app_manifest_create_error"
	ErrAppManifestGenerate                           = "app_manifest_generate_error"
	ErrAppManifestUpdate                             = "app_manifest_update_error"
	ErrAppManifestValidate                           = "app_manifest_validate_error"
	ErrAppNotEligible                                = "app_not_eligible"
	ErrAppNotInstalled                               = "app_not_installed"
	ErrAppNotFound                                   = "app_not_found"
	ErrAppNotHosted                                  = "app_not_hosted"
	ErrAppRemove                                     = "app_remove_error"
	ErrAppRenameApp                                  = "app_rename_app"
	ErrAuthProdTokenNotFound                         = "auth_prod_token_not_found"
	ErrAuthTimeout                                   = "auth_timeout_error"
	ErrAuthToken                                     = "auth_token_error"
	ErrAuthVerification                              = "auth_verification_error"
	ErrBotInviteRequired                             = "bot_invite_required" // Slack API error code
	ErrCannotAbandonApp                              = "cannot_abandon_app"
	ErrCannotAddOwner                                = "cannot_add_owner"
	ErrCannotCountOwners                             = "cannot_count_owners"
	ErrCannotDeleteApp                               = "cannot_delete_app"
	ErrCannotListCollaborators                       = "cannot_list_collaborators"
	ErrCannotListOwners                              = "cannot_list_owners"
	ErrCannotRemoveCollaborators                     = "cannot_remove_collaborators"
	ErrCannotRemoveOwners                            = "cannot_remove_owner"
	ErrCannotRevokeOrgBotToken                       = "cannot_revoke_org_bot_token"
	ErrConnectorApprovalPending                      = "connector_approval_pending"
	ErrConnectorApprovalRequired                     = "connector_approval_required"
	ErrConnectorDenied                               = "connector_denied"
	ErrConnectorNotInstalled                         = "connector_not_installed"
	ErrChannelNotFound                               = "channel_not_found"
	ErrCLIAutoUpdate                                 = "cli_autoupdate_error"
	ErrCLIConfigLocationError                        = "cli_config_location_error"
	ErrCLIConfigInvalid                              = "cli_config_invalid"
	ErrCLIReadError                                  = "cli_read_error"
	ErrCLIUpdateRequired                             = "cli_update_required" // Slack API error code
	ErrCommentRequired                               = "comment_required"
	ErrConnectedOrgDenied                            = "connected_org_denied"
	ErrConnectedTeamDenied                           = "connected_team_denied"
	ErrContextValueNotFound                          = "context_value_not_found"
	ErrCredentialsNotFound                           = "credentials_not_found"
	ErrCustomizableInputMissingMatchingWorkflowInput = "customizable_input_missing_matching_workflow_input"
	ErrCustomizableInputsNotAllowedOnOptionalInputs  = "customizable_inputs_not_allowed_on_optional_inputs"
	ErrCustomizableInputsOnlyAllowedOnLinkTriggers   = "customizable_inputs_only_allowed_on_link_triggers"
	ErrCustomizableInputUnsupportedType              = "customizable_input_unsupported_type"
	ErrDatastore                                     = "datastore_error"
	ErrDatastoreMissingPrimaryKey                    = "datastore_missing_primary_key"
	ErrDatastoreNotFound                             = "datastore_not_found"
	ErrDefaultAppAccess                              = "default_app_access_error"
	ErrDefaultAppSetting                             = "default_app_setting_error"
	ErrDenoNotFound                                  = "deno_not_found"
	ErrDeployedAppNotSupported                       = "deployed_app_not_supported"
	ErrDocumentationGenerationFailed                 = "documentation_generation_failed"
	ErrEnterpriseNotFound                            = "enterprise_not_found"
	ErrFailedAddingCollaborator                      = "failed_adding_collaborator"
	ErrFailedCreatingApp                             = "failed_creating_app"
	ErrFailedDatastoreOperation                      = "failed_datastore_operation"
	ErrFailedExport                                  = "failed_export"
	ErrFailForSomeRequests                           = "failed_for_some_requests"
	ErrFailedToGetUser                               = "failed_to_get_user"
	ErrFailedToSaveExtensionLogs                     = "failed_to_save_extension_logs"
	ErrFailToGetTeamsForRestrictedUser               = "fail_to_get_teams_for_restricted_user"
	ErrFeedbackNameInvalid                           = "feedback_name_invalid"
	ErrFeedbackNameRequired                          = "feedback_name_required"
	ErrFileRejected                                  = "file_rejected"
	ErrForbiddenTeam                                 = "forbidden_team"
	ErrFreeTeamNotAllowed                            = "free_team_not_allowed"
	ErrFunctionBelongsToAnotherApp                   = "function_belongs_to_another_app"
	ErrFunctionNotFound                              = "function_not_found"
	ErrGitNotFound                                   = "git_not_found"
	ErrGitClone                                      = "git_clone_error"
	ErrGitZipDownload                                = "git_zip_download_error"
	ErrHomeDirectoryAccessFailed                     = "home_directory_access_failed"
	ErrHooksJSONLocation                             = "hooks_json_location_error"
	ErrHostAppsDisallowUserScopes                    = "hosted_apps_disallow_user_scopes"
	ErrHTTPRequestFailed                             = "http_request_failed"
	ErrHTTPResponseInvalid                           = "http_response_invalid"
	ErrInsecureRequest                               = "insecure_request"
	ErrInstallationDenied                            = "installation_denied"
	ErrInstallationFailed                            = "installation_failed"
	ErrInstallationRequired                          = "installation_required"
	ErrInternal                                      = "internal_error"
	ErrInvalidApp                                    = "invalid_app"
	ErrInvalidAppDirectory                           = "invalid_app_directory"
	ErrInvalidAppFlag                                = "invalid_app_flag"
	ErrInvalidAppID                                  = "invalid_app_id"
	ErrInvalidArgs                                   = "invalid_args"
	ErrInvalidArgumentsCustomizableInputs            = "invalid_arguments_customizable_inputs"
	ErrInvalidArguments                              = "invalid_arguments"
	ErrInvalidAuth                                   = "invalid_auth"
	ErrInvalidChallenge                              = "invalid_challenge"
	ErrInvalidChannelID                              = "invalid_channel_id"
	ErrInvalidCursor                                 = "invalid_cursor"
	ErrInvalidDistributionType                       = "invalid_distribution_type"
	ErrInvalidFlag                                   = "invalid_flag"
	ErrInvalidInteractiveTriggerInputs               = "invalid_interactive_trigger_inputs"
	ErrInvalidManifest                               = "invalid_manifest"
	ErrInvalidManifestSource                         = "invalid_manifest_source"
	ErrInvalidParameters                             = "invalid_parameters"
	ErrInvalidPermissionType                         = "invalid_permission_type"
	ErrInvalidRefreshToken                           = "invalid_refresh_token"
	ErrInvalidRequestID                              = "invalid_request_id"
	ErrInvalidResourceID                             = "invalid_resource_id"
	ErrInvalidResourceType                           = "invalid_resource_type"
	ErrInvalidS3Key                                  = "invalid_s3_key"
	ErrInvalidScopes                                 = "invalid_scopes"
	ErrInvalidSemVer                                 = "invalid_semver"
	ErrInvalidSlackProjectDirectory                  = "invalid_slack_project_directory"
	ErrInvalidDatastore                              = "invalid_datastore"
	ErrInvalidDatastoreExpression                    = "invalid_datastore_expression"
	ErrInvalidToken                                  = "invalid_token"
	ErrInvalidTrigger                                = "invalid_trigger"
	ErrInvalidTriggerAccess                          = "invalid_trigger_access"
	ErrInvalidTriggerConfig                          = "invalid_trigger_config"
	ErrInvalidTriggerEventType                       = "invalid_trigger_event_type"
	ErrInvalidTriggerInputs                          = "invalid_trigger_inputs"
	ErrInvalidTriggerType                            = "invalid_trigger_type"
	ErrInvalidUserID                                 = "invalid_user_id"
	ErrInvalidWebhookConfig                          = "invalid_webhook_config"
	ErrInvalidWebhookSchemaRef                       = "invalid_webhook_schema_ref"
	ErrInvalidWorkflowAppID                          = "invalid_workflow_app_id"
	ErrInvalidWorkflowID                             = "invalid_workflow_id"
	ErrIsRestricted                                  = "is_restricted"
	ErrLocalAppNotFound                              = "local_app_not_found"
	ErrLocalAppNotSupported                          = "local_app_not_supported"
	ErrLocalAppRemoval                               = "local_app_removal_error"
	ErrLocalAppRun                                   = "local_app_run_error"
	ErrLocalAppRunCleanExit                          = "local_app_run_clean_exit"
	ErrMethodNotSupported                            = "method_not_supported"
	ErrMismatchedFlags                               = "mismatched_flags"
	ErrMissingAppID                                  = "missing_app_id"
	ErrMissingAppTeamID                              = "missing_app_team_id"
	ErrMissingChallenge                              = "missing_challenge"
	ErrMissingExperiment                             = "missing_experiment"
	ErrMissingExtension                              = "missing_extension"
	ErrMissingFunctionIdentifier                     = "missing_function_identifier"
	ErrMissingFlag                                   = "missing_flag"
	ErrMissingInput                                  = "missing_input"
	ErrMissingOptions                                = "missing_options"
	ErrMissingScope                                  = "missing_scope"
	ErrMissingScopes                                 = "missing_scopes"
	ErrMissingUser                                   = "missing_user"
	ErrMissingValue                                  = "missing_value"
	ErrNotAuthed                                     = "not_authed"
	ErrNotBearerToken                                = "not_bearer_token"
	ErrNotFound                                      = "not_found"
	ErrNoFile                                        = "no_file"
	ErrNoPendingRequest                              = "no_pending_request"
	ErrNoPermission                                  = "no_permission"
	ErrNoTokenFound                                  = "no_token_found"
	ErrNoTriggers                                    = "no_triggers"
	ErrNoValidNamedEntities                          = "no_valid_named_entities"
	ErrOrgNotConnected                               = "org_not_connected"
	ErrOrgNotFound                                   = "org_not_found"
	ErrOrgGrantExists                                = "org_grant_exists"
	ErrOSNotSupported                                = "os_not_supported"
	ErrOverResourceLimit                             = "over_resource_limit"
	ErrParameterValidationFailed                     = "parameter_validation_failed"
	ErrProcessInterrupted                            = "process_interrupted"
	ErrProjectCompilation                            = "project_compilation_error"
	ErrProjectConfigIDNotFound                       = "project_config_id_not_found"
	ErrProjectConfigManifestSource                   = "project_config_manifest_source_error"
	ErrProjectFileUpdate                             = "project_file_update_error"
	ErrProviderNotFound                              = "provider_not_found"
	ErrPrompt                                        = "prompt_error"
	ErrPublishedAppOnly                              = "published_app_only"
	ErrRequestIDOrAppIDIsRequired                    = "request_id_or_app_id_is_required"
	ErrRatelimited                                   = "ratelimited"
	ErrRestrictedPlanLevel                           = "restricted_plan_level"
	ErrRuntimeNotFound                               = "runtime_not_found"
	ErrRuntimeNotSupported                           = "runtime_not_supported"
	ErrSDKConfigLoad                                 = "sdk_config_load_error"
	ErrSDKHookInvocationFailed                       = "sdk_hook_invocation_failed"
	ErrSDKHookNotFound                               = "sdk_hook_not_found"
	ErrSampleCreate                                  = "sample_create_error"
	ErrServiceLimitsExceeded                         = "service_limits_exceeded"
	ErrSharedChannelDenied                           = "shared_channel_denied"
	ErrSlackAuth                                     = "slack_auth_error"
	ErrSlackJSONLocation                             = "slack_json_location_error"
	ErrSlackSlackJSONLocation                        = "slack_slack_json_location_error"
	ErrSocketConnection                              = "socket_connection_error"
	ErrScopesExceedAppConfig                         = "scopes_exceed_app_config"
	ErrStreamingActivityLogs                         = "streaming_activity_logs_error"
	ErrSurveyConfigNotFound                          = "survey_config_not_found"
	ErrSystemConfigIDNotFound                        = "system_config_id_not_found"
	ErrSystemRequirementsFailed                      = "system_requirements_failed"
	ErrTeamAccessNotGranted                          = "team_access_not_granted"
	ErrTeamFlagRequired                              = "team_flag_required"
	ErrTeamList                                      = "team_list_error"
	ErrTeamNotConnected                              = "team_not_connected"
	ErrTeamNotFound                                  = "team_not_found"
	ErrTeamNotOnEnterprise                           = "team_not_on_enterprise"
	ErrTeamQuotaExceeded                             = "team_quota_exceeded"
	ErrTemplatePathNotFound                          = "template_path_not_found"
	ErrTokenExpired                                  = "token_expired"
	ErrTokenRevoked                                  = "token_revoked"
	ErrTokenRotation                                 = "token_rotation_error"
	ErrTooManyCustomizableInputs                     = "too_many_customizable_inputs"
	ErrTooManyIdsProvided                            = "too_many_ids_provided"
	ErrTooManyNamedEntities                          = "too_many_named_entities"
	ErrTriggerCreate                                 = "trigger_create_error"
	ErrTriggerDelete                                 = "trigger_delete_error"
	ErrTriggerDoesNotExist                           = "trigger_does_not_exist"
	ErrTriggerNotFound                               = "trigger_not_found"
	ErrTriggerUpdate                                 = "trigger_update_error"
	ErrUnableToDelete                                = "unable_to_delete"
	ErrUnableToOpenFile                              = "unable_to_open_file"
	ErrUnableToParseJSON                             = "unable_to_parse_json"
	ErrUninstallHalted                               = "uninstall_halted"
	ErrUnknownFileType                               = "unknown_file_type"
	ErrUnknownFunctionID                             = "unknown_function_id"
	ErrUnknownMethod                                 = "unknown_method"
	ErrUnknownWebhookSchemaRef                       = "unknown_webhook_schema_ref"
	ErrUnknownWorkflowID                             = "unknown_workflow_id"
	ErrUntrustedSource                               = "untrusted_source"
	ErrUnsupportedFileName                           = "unsupported_file_name"
	ErrUserAlreadyOwner                              = "user_already_owner"
	ErrUserAlreadyRequested                          = "user_already_requested"
	ErrUserCannotManageApp                           = "user_cannot_manage_app"
	ErrUserIDIsRequired                              = "user_id_is_required"
	ErrUserNotFound                                  = "user_not_found"
	ErrUserRemovedFromTeam                           = "user_removed_from_team"
	ErrWorkflowNotFound                              = "workflow_not_found"
	ErrYaml                                          = "yaml_error"
)

var ErrorCodeMap = map[string]Error{

	ErrAccessDenied: {
		Code:        ErrAccessDenied,
		Message:     "You don't have the permission to access the specified resource",
		Remediation: "Check with your Slack admin to make sure that you have permission to access the resource.",
	},

	ErrAddAppToProject: {
		Code:    ErrAddAppToProject,
		Message: "Couldn't save your app's info to this project",
	},

	ErrAlreadyLoggedOut: {
		Code:    ErrAlreadyLoggedOut,
		Message: "You're already logged out",
	},

	ErrAlreadyResolved: {
		Code:    ErrAlreadyResolved,
		Message: "The app already has a resolution and cannot be requested",
	},

	ErrAppsList: {
		Code:    ErrAppsList,
		Message: "Couldn't get a list of your apps",
	},

	ErrAppAdd: {
		Code:    ErrAppAdd,
		Message: "Couldn't create a new app",
	},

	ErrAppApprovalRequestDenied: {
		Code:        ErrAppApprovalRequestDenied,
		Message:     "This app is currently denied for installation",
		Remediation: "Reach out to an admin for additional information, or try requesting again with different scopes and outgoing domains",
	},

	ErrAppApprovalRequestEligible: {
		Code:    ErrAppApprovalRequestEligible,
		Message: "This app requires permissions that must be reviewed by an admin before you can install it",
	},

	ErrAppApprovalRequestPending: {
		Code:        ErrAppApprovalRequestPending,
		Message:     "This app has requested admin approval to install and is awaiting review",
		Remediation: "Reach out to an admin for additional information",
	},

	ErrAppAuthTeamMismatch: {
		Code:        ErrAppAuthTeamMismatch,
		Message:     "Specified app and team are mismatched",
		Remediation: "Try a different combination of `--app` and `--team` flags",
	},

	ErrAppCreate: {
		Code:    ErrAppCreate,
		Message: "Couldn't create your app",
	},

	ErrAppDelete: {
		Code:    ErrAppDelete,
		Message: "Couldn't delete your app",
	},

	ErrAppDeploy: {
		Code:    ErrAppDeploy,
		Message: "Couldn't deploy your app",
	},

	ErrAppDeployNotSlackHosted: {
		Code:    ErrAppDeployNotSlackHosted,
		Message: "Deployment to Slack is not currently supported for apps with `runOnSlack` set as false",
		Details: ErrorDetails{
			ErrorDetail{Message: "Deployment to Slack is currently supported for apps written with the Deno Slack SDK."},
		},
		Remediation: fmt.Sprintf(`Learn about building apps with the Deno Slack SDK:

https://docs.slack.dev/tools/deno-slack-sdk

If you are using a Bolt framework, add a deploy hook then run: %s

Otherwise start your app for local development with: %s`,
			style.Commandf("deploy", true),
			style.Commandf("run", true),
		),
	},

	ErrAppDirectoryAccess: {
		Code:    ErrAppDirectoryAccess,
		Message: "Couldn't access app directory",
	},

	ErrAppDirOnlyFail: {
		Code:    ErrAppDirOnlyFail,
		Message: "The app was neither in the app directory nor created on this team/org, and cannot be requested",
	},

	ErrAppExists: {
		Code:    ErrAppExists,
		Message: "App already exists belonging to the team",
	},

	ErrAppFlagRequired: {
		Code:        ErrAppFlagRequired,
		Message:     "The --app flag must be provided",
		Remediation: "Choose a specific app with `--app <app_id>`",
	},

	ErrAppFound: {
		Code:    ErrAppFound,
		Message: "An app was found",
	},

	ErrAppHosted: {
		Code:    ErrAppHosted,
		Message: "App is configured for Run on Slack infrastructure",
	},

	ErrAppInstall: {
		Code:    ErrAppInstall,
		Message: "Couldn't install your app to a workspace",
	},

	ErrAppManifestAccess: {
		Code:    ErrAppManifestAccess,
		Message: "Couldn't access your app manifest",
	},

	ErrAppManifestCreate: {
		Code:    ErrAppManifestCreate,
		Message: "Couldn't create your app manifest",
	},

	ErrAppManifestGenerate: {
		Code:        ErrAppManifestGenerate,
		Message:     "Couldn't generate an app manifest from this project",
		Remediation: "Check to make sure you are in a valid Slack project directory and that your project has no compilation errors.",
	},

	ErrAppManifestUpdate: {
		Code:    ErrAppManifestUpdate,
		Message: "The app manifest was not updated",
	},

	ErrAppManifestValidate: {
		Code:    ErrAppManifestValidate,
		Message: "Your app manifest is invalid",
	},

	ErrAppNotEligible: {
		Code:    ErrAppNotEligible,
		Message: "The specified app is not eligible for this API",
	},

	ErrAppNotInstalled: {
		Code:    ErrAppNotInstalled,
		Message: "The provided app must be installed on this team",
	},

	ErrAppNotFound: {
		Code:    ErrAppNotFound,
		Message: "The app was not found",
	},

	ErrAppNotHosted: {
		Code:    ErrAppNotHosted,
		Message: "App is not configured to be deployed to the Slack platform",
		Remediation: strings.Join([]string{
			"Deploy an app containing workflow automations to Slack managed infrastructure",
			"Read about ROSI: https://docs.slack.dev/workflows/run-on-slack-infrastructure",
		}, "\n"),
	},

	ErrAppRemove: {
		Code:    ErrAppRemove,
		Message: "Couldn't remove your app",
	},

	ErrAppRenameApp: {
		Code:    ErrAppRenameApp,
		Message: "Couldn't rename your app",
	},

	ErrAuthProdTokenNotFound: {
		Code:    ErrAuthProdTokenNotFound,
		Message: "Couldn't find a valid auth token for the Slack API",
		Remediation: fmt.Sprintf(
			"You need to be logged in to at least 1 production (slack.com) team to use this command. Log into one with the %s command and try again.",
			style.Commandf("login", false),
		),
	},

	ErrAuthTimeout: {
		Code:        ErrAuthTimeout,
		Message:     "Couldn't receive authorization in the time allowed",
		Remediation: "Ensure you have pasted the command in a Slack workspace and accepted the permissions.",
	},

	ErrAuthToken: {
		Code:    ErrAuthToken,
		Message: "Couldn't get a token with an active session",
	},

	ErrCannotAbandonApp: {
		Code:    ErrCannotAbandonApp,
		Message: "The last owner cannot be removed",
	},

	ErrCannotAddOwner: {
		Code:    ErrCannotAddOwner,
		Message: "Unable to add the given user as owner",
	},

	ErrCannotCountOwners: {
		Code:    ErrCannotCountOwners,
		Message: "Unable to retrieve current app collaborators",
	},

	ErrConnectorApprovalPending: {
		Code:        ErrConnectorApprovalPending,
		Message:     "A connector requires admin approval before it can be installed\nApproval is pending review",
		Remediation: "Contact your Slack admin about the status of your request",
	},

	ErrConnectorApprovalRequired: {
		Code:        ErrConnectorApprovalRequired,
		Message:     "A connector requires admin approval before it can be installed",
		Remediation: "Request approval for the given connector from your Slack admin",
	},

	ErrConnectorDenied: {
		Code:        ErrConnectorDenied,
		Message:     "A connector has been denied for use by an admin",
		Remediation: "Contact your Slack admin",
	},

	ErrConnectorNotInstalled: {
		Code:        ErrConnectorNotInstalled,
		Message:     "A connector requires installation before it can be used",
		Remediation: "Request installation for the given connector",
	},

	ErrAuthVerification: {
		Code:    ErrAuthVerification,
		Message: "Couldn't verify your authorization",
	},

	ErrBotInviteRequired: {
		Code:        ErrBotInviteRequired,
		Message:     "Your app must be invited to the channel",
		Remediation: "Try to find the channel declared the source code of a workflow or function.\n\nOpen Slack, join the channel, invite your app, and try the command again.\nLearn more: https://slack.com/help/articles/201980108-Add-people-to-a-channel",
	},

	ErrCannotDeleteApp: {
		Code:    ErrCannotDeleteApp,
		Message: "Unable to delete app",
	},

	ErrCannotListCollaborators: {
		Code:    ErrCannotListCollaborators,
		Message: "Calling user is unable to list collaborators",
	},

	ErrCannotListOwners: {
		Code:    ErrCannotListOwners,
		Message: "Calling user is unable to list owners",
	},

	ErrCannotRemoveCollaborators: {
		Code:    ErrCannotRemoveCollaborators,
		Message: "User is unable to remove collaborators",
	},

	ErrCannotRemoveOwners: {
		Code:    ErrCannotRemoveOwners,
		Message: "Unable to remove the given user",
	},

	ErrCannotRevokeOrgBotToken: {
		Code:    ErrCannotRevokeOrgBotToken,
		Message: "Revoking org-level bot token is not supported",
	},

	ErrChannelNotFound: {
		Code:        ErrChannelNotFound,
		Message:     "Couldn't find the specified Slack channel",
		Remediation: "Try adding your app as a member to the channel.",
	},

	ErrCLIAutoUpdate: {
		Code:        ErrCLIAutoUpdate,
		Message:     "Couldn't auto-update this command-line tool",
		Remediation: "You can manually install the latest version from:\nhttps://docs.slack.dev/tools/slack-cli",
	},

	ErrCLIConfigLocationError: {
		Code:    ErrCLIConfigLocationError,
		Message: fmt.Sprintf("The %s configuration file is not supported", filepath.Join(".slack", "cli.json")),
		Remediation: strings.Join([]string{
			"This version of the CLI no longer supports this configuration file.",
			fmt.Sprintf("Move the %s file to %s and try again.", filepath.Join(".slack", "cli.json"), filepath.Join(".slack", "hooks.json")),
		}, "\n"),
	},

	ErrCLIReadError: {
		Code:        ErrCLIReadError,
		Message:     "There was an error reading configuration",
		Remediation: "Check your config.json file.",
	},

	ErrCLIConfigInvalid: {
		Code:        ErrCLIConfigInvalid,
		Message:     "Configuration invalid",
		Remediation: "Check your config.json file.",
	},

	ErrCLIUpdateRequired: {
		Code:        ErrCLIUpdateRequired,
		Message:     "Slack API requires the latest version of the Slack CLI",
		Remediation: fmt.Sprintf("You can upgrade to the latest version of the Slack CLI using the command: %s", style.Commandf("upgrade", false)),
	},

	ErrConnectedOrgDenied: {
		Code:    ErrConnectedOrgDenied,
		Message: "The admin does not allow connected organizations to be named_entities",
	},

	ErrCommentRequired: {
		Code:    ErrCommentRequired,
		Message: "Your admin is requesting a reason to approve installation of this app",
	},

	ErrConnectedTeamDenied: {
		Code:    ErrConnectedTeamDenied,
		Message: "The admin does not allow connected teams to be named_entities",
	},

	ErrContextValueNotFound: {
		Code:    ErrContextValueNotFound,
		Message: "The context value could not be found",
	},

	ErrCredentialsNotFound: {
		Code:        ErrCredentialsNotFound,
		Message:     "No authentication found for this team",
		Remediation: fmt.Sprintf("Use the command %s to login to this workspace", style.Commandf("login", false)),
	},

	ErrCustomizableInputMissingMatchingWorkflowInput: {
		Code:    ErrCustomizableInputMissingMatchingWorkflowInput,
		Message: "Customizable input on the trigger must map to a workflow input of the same name",
	},

	ErrCustomizableInputsNotAllowedOnOptionalInputs: {
		Code:    ErrCustomizableInputsNotAllowedOnOptionalInputs,
		Message: "Customizable trigger inputs must map to required workflow inputs",
	},

	ErrCustomizableInputsOnlyAllowedOnLinkTriggers: {
		Code:    ErrCustomizableInputsOnlyAllowedOnLinkTriggers,
		Message: "Customizable inputs are only allowed on link triggers",
	},

	ErrCustomizableInputUnsupportedType: {
		Code:    ErrCustomizableInputUnsupportedType,
		Message: "Customizable input has been mapped to a workflow input of an unsupported type. Only `UserID`, `ChannelId`, and `String` are supported for customizable inputs",
	},

	ErrDatastore: {
		Code:    ErrDatastore,
		Message: "An error occurred while accessing your datastore",
	},

	ErrDatastoreMissingPrimaryKey: {
		Code:    ErrDatastoreMissingPrimaryKey,
		Message: "The primary key for the datastore is missing",
	},

	ErrDatastoreNotFound: {
		Code:    ErrDatastoreNotFound,
		Message: "The specified datastore could not be found",
	},

	ErrDefaultAppAccess: {
		Code:    ErrDefaultAppAccess,
		Message: "Couldn't access the default app",
	},

	ErrDefaultAppSetting: {
		Code:    ErrDefaultAppSetting,
		Message: "Couldn't set this app as the default",
	},

	ErrDenoNotFound: {
		Code:        ErrDenoNotFound,
		Message:     "Couldn't find the 'deno' language runtime installed on this system",
		Remediation: "To install Deno, visit https://deno.land/#installation.",
	},

	ErrDeployedAppNotSupported: {
		Code:    ErrDeployedAppNotSupported,
		Message: "A deployed app cannot be used by this command",
	},

	ErrDocumentationGenerationFailed: {
		Code:    ErrDocumentationGenerationFailed,
		Message: "Failed to generate documentation",
	},

	ErrEnterpriseNotFound: {
		Code:    ErrEnterpriseNotFound,
		Message: "The `enterprise` was not found",
	},

	ErrFailedAddingCollaborator: {
		Code:    ErrFailedAddingCollaborator,
		Message: "Failed writing a collaborator record for this new app",
	},

	ErrFailedCreatingApp: {
		Code:    ErrFailedCreatingApp,
		Message: "Failed to create the app model",
	},

	ErrFailedDatastoreOperation: {
		Code:        ErrFailedDatastoreOperation,
		Message:     "Failed while managing datastore infrastructure",
		Remediation: "Please try again and reach out to feedback@slack.com if the problem persists.",
	},

	ErrFailedExport: {
		Code:    ErrFailedExport,
		Message: "Couldn't export the app manifest",
	},

	ErrFailForSomeRequests: {
		Code:    ErrFailForSomeRequests,
		Message: "At least one request was not cancelled",
	},

	ErrFailedToGetUser: {
		Code:    ErrFailedToGetUser,
		Message: "Couldn't find the user to install the app",
	},

	ErrFailedToSaveExtensionLogs: {
		Code:    ErrFailedToSaveExtensionLogs,
		Message: "Couldn't save the logs",
	},

	ErrFailToGetTeamsForRestrictedUser: {
		Code:    ErrFailToGetTeamsForRestrictedUser,
		Message: "Failed to get teams for restricted user",
	},

	ErrFeedbackNameInvalid: {
		Code:        ErrFeedbackNameInvalid,
		Message:     "The name of the feedback is invalid",
		Remediation: fmt.Sprintf("View the feedback options with %s", style.Commandf("feedback --help", false)),
	},

	ErrFeedbackNameRequired: {
		Code:    ErrFeedbackNameRequired,
		Message: "The name of the feedback is required",
		Remediation: strings.Join([]string{
			"Please provide a `--name <string>` flag or remove the `--no-prompt` flag",
			fmt.Sprintf("View feedback options with %s", style.Commandf("feedback --help", false)),
		}, "\n"),
	},

	ErrFileRejected: {
		Code:    ErrFileRejected,
		Message: "Not an acceptable S3 file",
	},

	ErrForbiddenTeam: {
		Code:    ErrForbiddenTeam,
		Message: "The authenticated team cannot use this API",
	},

	ErrFreeTeamNotAllowed: {
		Code:        ErrFreeTeamNotAllowed,
		Message:     "Free workspaces do not support the Slack platform's low-code automation for workflows and functions",
		Remediation: "You can install this app if you upgrade your workspace: https://slack.com/pricing.",
	},

	ErrFunctionBelongsToAnotherApp: {
		Code:    ErrFunctionBelongsToAnotherApp,
		Message: "The provided function_id does not belong to this app_id",
	},

	ErrFunctionNotFound: {
		Code:    ErrFunctionNotFound,
		Message: "The specified function couldn't be found",
	},

	ErrGitNotFound: {
		Code:        ErrGitNotFound,
		Message:     "Couldn't find Git installed on this system",
		Remediation: "To install Git, visit https://github.com/git-guides/install-git.",
	},

	ErrGitClone: {
		Code:    ErrGitClone,
		Message: "Git failed to clone repository",
	},

	ErrGitZipDownload: {
		Code:    ErrGitZipDownload,
		Message: "Cannot download Git repository as a .zip archive",
	},

	ErrHomeDirectoryAccessFailed: {
		Code:        ErrHomeDirectoryAccessFailed,
		Message:     "Failed to read/create .slack/ directory in your home directory",
		Remediation: "A Slack directory is required for retrieving/storing auth credentials and config data. Check permissions on your system.",
	},

	ErrHooksJSONLocation: {
		Code:        ErrHooksJSONLocation,
		Message:     "Missing the Slack hooks file from project configurations",
		Remediation: fmt.Sprintf("A `%s` file must be present in the project's `.slack` directory.", filepath.Join(".slack", "hooks.json")),
	},

	ErrHostAppsDisallowUserScopes: {
		Code:    ErrHostAppsDisallowUserScopes,
		Message: "Hosted apps do not support user scopes",
	},

	ErrHTTPRequestFailed: {
		Code:    ErrHTTPRequestFailed,
		Message: "HTTP request failed",
	},

	ErrHTTPResponseInvalid: {
		Code:    ErrHTTPResponseInvalid,
		Message: "Received an invalid response from the server",
	},

	ErrInsecureRequest: {
		Code:    ErrInsecureRequest,
		Message: "The method was not called via a `POST` request",
	},

	ErrInstallationDenied: {
		Code:        ErrInstallationDenied,
		Message:     "Couldn't install the app because the installation request was denied",
		Remediation: "Reach out to one of your App Managers for additional information.",
	},

	ErrInstallationFailed: {
		Code:    ErrInstallationFailed,
		Message: "Couldn't install the app",
	},

	ErrInstallationRequired: {
		Code:        ErrInstallationRequired,
		Message:     "A valid installation of this app is required to take this action",
		Remediation: fmt.Sprintf("Install the app with %s", style.Commandf("install", false)),
	},

	ErrInternal: {
		Code:        ErrInternal,
		Message:     "An internal error has occurred with the Slack platform",
		Remediation: "Please reach out to feedback@slack.com if the problem persists.",
	},

	ErrInvalidApp: {
		Code:    ErrInvalidApp,
		Message: "Either the app does not exist or an app created from the provided manifest would not be valid",
	},

	ErrInvalidAppDirectory: {
		Code:    ErrInvalidAppDirectory,
		Message: "This is an invalid Slack app project directory",
		Remediation: strings.Join([]string{
			fmt.Sprintf("A valid Slack project includes the Slack hooks file: %s", filepath.Join(".slack", "hooks.json")),
			"",
			"If this is a Slack project, you can initialize it with " + style.Commandf("init", false),
		}, "\n"),
	},

	ErrInvalidAppFlag: {
		Code:        ErrInvalidAppFlag,
		Message:     "The provided --app flag value is not valid",
		Remediation: "Specify the environment with `--app local` or `--app deployed`\nOr choose a specific app with `--app <app_id>`",
	},

	ErrInvalidAppID: {
		Code:        ErrInvalidAppID,
		Message:     "App ID may be invalid for this user account and workspace",
		Remediation: "Check to make sure you are signed into the correct workspace for this app and you have the required permissions to perform this action.",
	},

	ErrInvalidArgs: {
		Code:    ErrInvalidArgs,
		Message: "Required arguments either were not provided or contain invalid values",
	},

	ErrInvalidArgumentsCustomizableInputs: {
		Code:    ErrInvalidArgumentsCustomizableInputs,
		Message: "A trigger input parameter with customizable: true cannot be set as hidden or locked, nor have a value provided at trigger creation time",
	},

	ErrInvalidArguments: {
		Code:    ErrInvalidArguments,
		Message: "Slack API request parameters are invalid",
	},

	ErrInvalidAuth: {
		Code:    ErrInvalidAuth,
		Message: "Your user account authorization isn't valid",
		Remediation: fmt.Sprintf(
			"Your user account authorization may be expired or does not have permission to access the resource. Try to login to the same user account again using %s.",
			style.Commandf("login", false),
		),
	},

	ErrInvalidChallenge: {
		Code:        ErrInvalidChallenge,
		Message:     "The challenge code is invalid",
		Remediation: fmt.Sprintf("The previous slash command and challenge code have now expired. To retry, use %s, paste the slash command in any Slack channel, and enter the challenge code displayed by Slack. It is easiest to copy & paste the challenge code.", style.Commandf("login", false)),
	},

	ErrInvalidChannelID: {
		Code:        ErrInvalidChannelID,
		Message:     "Channel ID specified doesn't exist or you do not have permissions to access it",
		Remediation: "Channel ID appears to be formatted correctly. Check if this channel exists on the current team and that you have permissions to access it.",
	},

	ErrInvalidCursor: {
		Code:    ErrInvalidCursor,
		Message: "Value passed for `cursor` was not valid or is valid no longer",
	},

	ErrInvalidFlag: {
		Code:    ErrInvalidFlag,
		Message: "The provided flag value is invalid",
	},

	ErrInvalidDistributionType: {
		Code:    ErrInvalidDistributionType,
		Message: "This function requires distribution_type to be set as named_entities before adding users",
	},

	ErrInvalidInteractiveTriggerInputs: {
		Code:    ErrInvalidInteractiveTriggerInputs,
		Message: "One or more input parameter types isn't supported by the link trigger type",
	},

	ErrInvalidManifest: {
		Code:    ErrInvalidManifest,
		Message: "The provided manifest file does not validate against schema. Consult the additional errors field to locate specific issues",
	},

	ErrInvalidManifestSource: {
		Code:    ErrInvalidManifestSource,
		Message: "A manifest does not exist at the provided source",
		Remediation: strings.Join([]string{
			fmt.Sprintf("Set 'manifest.source' to either \"remote\" or \"local\" in %s", filepath.Join(".slack", "config.json")),
			fmt.Sprintf("Read about manifest sourcing with the %s command", style.Commandf("manifest info --help", false)),
		}, "\n"),
	},

	ErrInvalidParameters: {
		Code:    ErrInvalidParameters,
		Message: "slack_cli_version supplied is invalid",
	},

	ErrInvalidPermissionType: {
		Code:    ErrInvalidPermissionType,
		Message: "Permission type must be set to `named_entities` before you can manage users",
	},

	ErrInvalidRefreshToken: {
		Code:    ErrInvalidRefreshToken,
		Message: "The given refresh token is invalid",
	},

	ErrInvalidRequestID: {
		Code:    ErrInvalidRequestID,
		Message: "The request_id passed is invalid",
	},

	ErrInvalidResourceID: {
		Code:    ErrInvalidResourceID,
		Message: "The resource_id for the given resource_type is invalid",
	},

	ErrInvalidResourceType: {
		Code:    ErrInvalidResourceType,
		Message: "The resource_type argument is invalid.",
	},

	ErrInvalidS3Key: {
		Code:        ErrInvalidS3Key,
		Message:     "An internal error occurred",
		Remediation: "Please reach out to feedback@slack.com if the problem persists.",
	},

	ErrInvalidScopes: {
		Code:    ErrInvalidScopes,
		Message: "Some of the provided scopes do not exist",
	},

	ErrInvalidSemVer: {
		Code:    ErrInvalidSemVer,
		Message: "The provided version does not follow semantic versioning",
	},

	ErrInvalidSlackProjectDirectory: {
		Code:        ErrInvalidSlackProjectDirectory,
		Message:     "Current directory is not a Slack project",
		Remediation: fmt.Sprintf("Change in to a Slack project directory. A Slack project always includes the Slack hooks file (`%s`).", filepath.Join(".slack", "hooks.json")),
	},

	ErrInvalidDatastore: {
		Code:    ErrInvalidDatastore,
		Message: "Invalid datastore specified in your project",
	},

	ErrInvalidDatastoreExpression: {
		Code:    ErrInvalidDatastoreExpression,
		Message: "The provided expression is not valid",
		Remediation: strings.Join([]string{
			"Verify the expression you provided is valid JSON surrounded by quotations",
			fmt.Sprintf("Use %s for examples", style.Commandf("datastore --help", false)),
		}, "\n"),
	},

	ErrInvalidToken: {
		Code:    ErrInvalidToken,
		Message: "The provided token is not valid",
	},

	ErrInvalidTrigger: {
		Code:    ErrInvalidTrigger,
		Message: "Invalid trigger specified in your project",
	},

	ErrInvalidTriggerAccess: {
		Code:    ErrInvalidTriggerAccess,
		Message: "Trigger access can not be configured for more than 10 users",
	},

	ErrInvalidTriggerConfig: {
		Code:    ErrInvalidTriggerConfig,
		Message: "The provided trigger object does not conform to the trigger type's schema",
	},

	ErrInvalidTriggerEventType: {
		Code:    ErrInvalidTriggerEventType,
		Message: "The provided event type is not allowed",
	},

	ErrInvalidTriggerInputs: {
		Code:    ErrInvalidTriggerInputs,
		Message: "Required inputs for the referenced function/workflow are not passed",
	},

	ErrInvalidTriggerType: {
		Code:    ErrInvalidTriggerType,
		Message: "The provided trigger type is not recognized",
	},

	ErrInvalidUserID: {
		Code:    ErrInvalidUserID,
		Message: "A value passed as a user_id is invalid",
	},

	ErrInvalidWebhookConfig: {
		Code:    ErrInvalidWebhookConfig,
		Message: "Only one of schema or schema_ref should be provided",
	},

	ErrInvalidWebhookSchemaRef: {
		Code:    ErrInvalidWebhookSchemaRef,
		Message: "Unable to parse the schema ref",
	},

	ErrInvalidWorkflowAppID: {
		Code:    ErrInvalidWorkflowAppID,
		Message: "A value passed as workflow_app_id is invalid or missing",
	},

	ErrInvalidWorkflowID: {
		Code:    ErrInvalidWorkflowID,
		Message: "A value passed as a workflow ID is invalid",
	},

	ErrIsRestricted: {
		Code:    ErrIsRestricted,
		Message: "Restricted users cannot request",
	},

	ErrLocalAppNotFound: {
		Code:    ErrLocalAppNotFound,
		Message: "Couldn't find the local app",
	},

	ErrLocalAppNotSupported: {
		Code:    ErrLocalAppNotSupported,
		Message: "A local app cannot be used by this command",
	},

	ErrLocalAppRemoval: {
		Code:    ErrLocalAppRemoval,
		Message: "Couldn't remove local app",
	},

	ErrLocalAppRun: {
		Code:    ErrLocalAppRun,
		Message: "Couldn't run app locally",
	},

	ErrMethodNotSupported: {
		Code:    ErrMethodNotSupported,
		Message: "This API method is not supported",
	},

	ErrMismatchedFlags: {
		Code:    ErrMismatchedFlags,
		Message: "The provided flags cannot be used together",
	},

	ErrMissingAppID: {
		Code:    ErrMissingAppID,
		Message: "workflow_app_id is required to update via workflow reference",
	},

	ErrMissingAppTeamID: {
		Code:    ErrMissingAppTeamID,
		Message: "team_id is required to create or update this app",
	},

	ErrMissingChallenge: {
		Code:    ErrMissingChallenge,
		Message: "Challenge must be supplied",
	},

	ErrMissingExperiment: {
		Code:    ErrMissingExperiment,
		Message: "The feature is behind an experiment not toggled on",
	},

	ErrMissingExtension: {
		Code:    ErrMissingExtension,
		Message: "An extension is missing",
	},

	ErrMissingFunctionIdentifier: {
		Code:    ErrMissingFunctionIdentifier,
		Message: "Could not find the given workflow using the specified reference",
	},

	ErrMissingFlag: {
		Code:    ErrMissingFlag,
		Message: "An argument must be provided for the flag",
	},

	ErrMissingInput: {
		Code:    ErrMissingInput,
		Message: "A required value was not supplied as input",
	},

	ErrMissingOptions: {
		Code:    ErrMissingOptions,
		Message: "There are no options to select from",
	},

	ErrMissingScope: {
		Code:    ErrMissingScope,
		Message: "Your login is out of date",
		Remediation: fmt.Sprintf(
			"Run %s and then %s again.",
			style.Commandf("logout", false),
			style.Commandf("login", false),
		),
	},

	ErrMissingScopes: {
		Code:    ErrMissingScopes,
		Message: "Additional scopes are required to create this type of trigger",
	},

	ErrMissingUser: {
		Code:    ErrMissingUser,
		Message: "The `user` was not found",
	},

	ErrMissingValue: {
		Code:    ErrMissingValue,
		Message: "Missing `value` property on an input. You must either provide the value now, or mark this input as `customizable`: `true` and provide the value at the time the trigger is executed.",
	},

	ErrNotAuthed: {
		Code:        ErrNotAuthed,
		Message:     "You are either not logged in or your login session has expired",
		Remediation: fmt.Sprintf("Authorize your CLI with %s", style.Commandf("login", false)),
	},

	ErrNotBearerToken: {
		Code:    ErrNotBearerToken,
		Message: "Incompatible token type provided",
	},

	ErrNotFound: {
		Code:    ErrNotFound,
		Message: "Couldn't find row",
	},

	ErrNoFile: {
		Code:        ErrNoFile,
		Message:     "Couldn't upload your bundled code to server",
		Remediation: "Please try again",
	},

	ErrNoPendingRequest: {
		Code:    ErrNoPendingRequest,
		Message: "Pending request not found",
	},

	ErrNoPermission: {
		Code:        ErrNoPermission,
		Message:     "You are either not a collaborator on this app or you do not have permissions to perform this action",
		Remediation: "Contact the app owner to add you as a collaborator",
	},

	ErrNoTokenFound: {
		Code:    ErrNoTokenFound,
		Message: "No tokens found to delete",
	},

	ErrNoTriggers: {
		Code:    ErrNoTriggers,
		Message: "There are no triggers installed for this app",
	},

	ErrNoValidNamedEntities: {
		Code:    ErrNoValidNamedEntities,
		Message: "None of the provided named entities were valid",
	},

	ErrOrgNotConnected: {
		Code:    ErrOrgNotConnected,
		Message: "One or more of the listed organizations was not connected",
	},

	ErrOrgNotFound: {
		Code:    ErrOrgNotFound,
		Message: "One or more of the listed organizations could not be found",
	},

	ErrOrgGrantExists: {
		Code:    ErrOrgGrantExists,
		Message: "A different org workspace grant already exists for the installed app",
	},

	ErrOSNotSupported: {
		Code:    ErrOSNotSupported,
		Message: "This operating system is not supported",
	},

	ErrOverResourceLimit: {
		Code:    ErrOverResourceLimit,
		Message: "Workspace exceeded the maximum number of Run On Slack functions and/or app datastores.",
	},

	ErrParameterValidationFailed: {
		Code:    ErrParameterValidationFailed,
		Message: "There were problems when validating the inputs against the function parameters. See API response for more details",
	},

	ErrProcessInterrupted: {
		Code:    ErrProcessInterrupted,
		Message: "The process received an interrupt signal",
	},

	ErrProjectCompilation: {
		Code:    ErrProjectCompilation,
		Message: "An error occurred while compiling your code",
	},

	ErrProjectConfigIDNotFound: {
		Code:    ErrProjectConfigIDNotFound,
		Message: `The "project_id" property is missing from the project-level configuration file`,
	},

	ErrProjectConfigManifestSource: {
		Code:    ErrProjectConfigManifestSource,
		Message: "Project manifest source is not valid",
		Remediation: strings.Join([]string{
			fmt.Sprintf("Set 'manifest.source' to either \"remote\" or \"local\" in %s", filepath.Join(".slack", "config.json")),
			fmt.Sprintf("Read about manifest sourcing with the %s command", style.Commandf("manifest info --help", false)),
		}, "\n"),
	},

	ErrProjectFileUpdate: {
		Code:    ErrProjectFileUpdate,
		Message: "Failed to update project files",
	},

	ErrProviderNotFound: {
		Code:    ErrProviderNotFound,
		Message: "The provided provider_key is invalid",
	},

	ErrPrompt: {
		Code:    ErrPrompt,
		Message: "An error occurred while executing prompts",
	},

	ErrPublishedAppOnly: {
		Code:    ErrPublishedAppOnly,
		Message: "This action is only permitted for published app IDs",
	},

	ErrRequestIDOrAppIDIsRequired: {
		Code:    ErrRequestIDOrAppIDIsRequired,
		Message: "Must include a request_id or app_id",
	},

	ErrRatelimited: {
		Code:    ErrRatelimited,
		Message: "Too many calls in succession during a short period of time",
	},

	ErrRestrictedPlanLevel: {
		Code:    ErrRestrictedPlanLevel,
		Message: "Your Slack plan does not have access to the requested feature",
	},

	ErrRuntimeNotFound: {
		Code:        ErrRuntimeNotFound,
		Message:     "The hook runtime executable was not found",
		Remediation: "Make sure the required runtime has been installed to run hook scripts.",
	},

	ErrRuntimeNotSupported: {
		Code:    ErrRuntimeNotSupported,
		Message: "The SDK runtime is not supported by the CLI",
	},

	ErrSampleCreate: {
		Code:    ErrSampleCreate,
		Message: "Couldn't create app from sample",
	},

	ErrServiceLimitsExceeded: {
		Code:    ErrServiceLimitsExceeded,
		Message: "Your workspace has exhausted the 10 apps limit for free teams. To create more apps, upgrade your Slack plan at https://my.slack.com/plans",
	},

	ErrSharedChannelDenied: {
		Code:    ErrSharedChannelDenied,
		Message: "The team admin does not allow shared channels to be named_entities",
	},

	ErrSDKConfigLoad: {
		Code:        ErrSDKConfigLoad,
		Message:     fmt.Sprintf("There was an error while reading the Slack hooks file (`%s`) or running the `get-hooks` hook", filepath.Join(".slack", "hooks.json")),
		Remediation: fmt.Sprintf("Run %s to check that your system dependencies are up-to-date.", style.Commandf("doctor", false)),
	},

	ErrSDKHookInvocationFailed: {
		Code:        ErrSDKHookInvocationFailed,
		Message:     fmt.Sprintf("A script hook defined in the Slack Configuration file (`%s`) returned an error", filepath.Join(".slack", "hooks.json")),
		Remediation: fmt.Sprintf("Run %s to check that your system dependencies are up-to-date.", style.Commandf("doctor", false)),
	},

	ErrSDKHookNotFound: {
		Code:    ErrSDKHookNotFound,
		Message: fmt.Sprintf("A script in %s was not found", style.Highlight(filepath.Join(".slack", "hooks.json"))),
		Remediation: strings.Join([]string{
			"Hook scripts are defined in one of these Slack hooks files:",
			"- slack.json",
			"- " + filepath.Join(".slack", "hooks.json"),
			"",
			"Every app requires a Slack hooks file and you can find an example at:",
			style.Highlight("https://github.com/slack-samples/deno-starter-template/blob/main/.slack/hooks.json"),
			"",
			"You can create a hooks file manually or with the " + style.Commandf("init", false) + " command.",
			"",
			"When manually creating the hooks file, you must install the hook dependencies.",
		}, "\n"),
	},

	ErrSlackAuth: {
		Code:        ErrSlackAuth,
		Message:     "You are not logged into a team or have not installed an app",
		Remediation: fmt.Sprintf("Use the command %s to login and %s to install your app", style.Commandf("login", false), style.Commandf("install", false)),
	},

	ErrSlackJSONLocation: {
		Code:    ErrSlackJSONLocation,
		Message: "The slack.json configuration file is deprecated",
		Remediation: strings.Join([]string{
			"Next major version of the CLI will no longer support this configuration file.",
			fmt.Sprintf("Move the slack.json file to %s and continue onwards.", filepath.Join(".slack", "hooks.json")),
		}, "\n"),
	},

	ErrSlackSlackJSONLocation: {
		Code:    ErrSlackSlackJSONLocation,
		Message: fmt.Sprintf("The %s configuration file is deprecated", filepath.Join(".slack", "slack.json")),
		Remediation: strings.Join([]string{
			"Next major version of the CLI will no longer support this configuration file.",
			fmt.Sprintf("Move the %s file to %s and proceed again.", filepath.Join(".slack", "slack.json"), filepath.Join(".slack", "hooks.json")),
		}, "\n"),
	},

	ErrSocketConnection: {
		Code:    ErrSocketConnection,
		Message: "Couldn't connect to Slack over WebSocket",
	},

	ErrScopesExceedAppConfig: {
		Code:    ErrScopesExceedAppConfig,
		Message: "Scopes requested exceed app configuration",
	},

	ErrStreamingActivityLogs: {
		Code:    "streaming_activity_logs_error",
		Message: "Failed to stream the most recent activity logs",
	},

	ErrSurveyConfigNotFound: {
		Code:    ErrSurveyConfigNotFound,
		Message: "Survey config not found",
	},

	ErrSystemConfigIDNotFound: {
		Code:    ErrSystemConfigIDNotFound,
		Message: `The "system_id" property is missing from the system-level configuration file`,
	},

	ErrSystemRequirementsFailed: {
		Code:    ErrSystemRequirementsFailed,
		Message: "Couldn't verify all system requirements",
	},

	ErrTeamAccessNotGranted: {
		Code:    ErrTeamAccessNotGranted,
		Message: "There was an issue granting access to the team",
	},

	ErrTeamFlagRequired: {
		Code:        ErrTeamFlagRequired,
		Message:     "The --team flag must be provided",
		Remediation: "Choose a specific team with `--team <team_domain>` or `--team <team_id>`",
	},

	ErrTeamList: {
		Code:    ErrTeamList,
		Message: "Couldn't get a list of teams",
	},

	ErrTeamNotConnected: {
		Code:    ErrTeamNotConnected,
		Message: "One or more of the listed teams was not connected by org",
	},

	ErrTeamNotFound: {
		Code:    ErrTeamNotFound,
		Message: "Team could not be found",
	},

	ErrTeamNotOnEnterprise: {
		Code:    ErrTeamNotOnEnterprise,
		Message: "Cannot query team by domain because team is not on an enterprise",
	},

	ErrTeamQuotaExceeded: {
		Code:    ErrTeamQuotaExceeded,
		Message: "Total number of requests exceeded team quota",
	},

	ErrTemplatePathNotFound: {
		Code:    "template_path_not_found",
		Message: "No template app was found at the provided path",
	},

	ErrTokenExpired: {
		Code:        ErrTokenExpired,
		Message:     "Your access token has expired",
		Remediation: fmt.Sprintf("Use the command %s to authenticate again", style.Commandf("login", false)),
	},

	ErrTokenRevoked: {
		Code:        ErrTokenRevoked,
		Message:     "Your token has already been revoked",
		Remediation: fmt.Sprintf("Use the command %s to authenticate again", style.Commandf("login", false)),
	},

	ErrTokenRotation: {
		Code:        ErrTokenRotation,
		Message:     "An error occurred while rotating your access token",
		Remediation: fmt.Sprintf("Use the command %s to authenticate again", style.Commandf("login", false)),
	},

	ErrTooManyCustomizableInputs: {
		Code:    ErrTooManyCustomizableInputs,
		Message: "Cannot have more than 10 customizable inputs",
	},

	ErrTooManyIdsProvided: {
		Code:    ErrTooManyIdsProvided,
		Message: "Ensure you provide only app_id OR request_id",
	},

	ErrTooManyNamedEntities: {
		Code:    ErrTooManyNamedEntities,
		Message: "Too many named entities passed into the trigger permissions setting",
	},

	ErrTriggerCreate: {
		Code:    ErrTriggerCreate,
		Message: "Couldn't create a trigger",
	},

	ErrTriggerDelete: {
		Code:    ErrTriggerDelete,
		Message: "Couldn't delete a trigger",
	},

	ErrTriggerDoesNotExist: {
		Code:    ErrTriggerDoesNotExist,
		Message: "The trigger provided does not exist",
	},

	ErrTriggerNotFound: {
		Code:    ErrTriggerNotFound,
		Message: "The specified trigger cannot be found",
	},

	ErrTriggerUpdate: {
		Code:    ErrTriggerUpdate,
		Message: "Couldn't update a trigger",
	},

	ErrUnableToDelete: {
		Code:    ErrUnableToDelete,
		Message: "There was an error deleting tokens",
	},

	ErrUnableToOpenFile: {
		Code:    ErrUnableToOpenFile,
		Message: "Error with file upload",
	},

	ErrUnableToParseJSON: {
		Code:    ErrUnableToParseJSON,
		Message: "`<json>` Couldn't be parsed as a json object",
	},

	ErrUninstallHalted: {
		Code:    ErrUninstallHalted,
		Message: "The uninstall process was interrupted",
	},

	ErrUnknownFileType: {
		Code:    ErrUnknownFileType,
		Message: "Unknown file type, must be application/zip",
	},

	ErrUnknownFunctionID: {
		Code:    ErrUnknownFunctionID,
		Message: "The provided function_id was not found",
	},

	ErrUnknownMethod: {
		Code:    ErrUnknownMethod,
		Message: "The Slack API method does not exist or you do not have permissions to access it",
	},

	ErrUnknownWebhookSchemaRef: {
		Code:    ErrUnknownWebhookSchemaRef,
		Message: "Unable to find the corresponding type based on the schema ref",
	},

	ErrUntrustedSource: {
		Code:        ErrUntrustedSource,
		Message:     "Source is by an unknown or untrusted author",
		Remediation: "Use --force flag or set trust_unknown_sources: true in config.json file to disable warning",
	},

	ErrUnknownWorkflowID: {
		Code:    ErrUnknownWorkflowID,
		Message: "The provided workflow_id was not found for this app",
	},

	ErrUserIDIsRequired: {
		Code:    ErrUserIDIsRequired,
		Message: "Must include a user_id to cancel request for an app with app_id",
	},

	ErrUnsupportedFileName: {
		Code:    ErrUnsupportedFileName,
		Message: "File name is not supported",
	},

	ErrUserAlreadyOwner: {
		Code:    ErrUserAlreadyOwner,
		Message: "The user is already an owner for this app",
	},

	ErrUserAlreadyRequested: {
		Code:    ErrUserAlreadyRequested,
		Message: "The user has a request pending for this app",
	},

	ErrUserCannotManageApp: {
		Code:        ErrUserCannotManageApp,
		Message:     "You do not have permissions to install this app",
		Remediation: "Reach out to one of your App Managers to request permissions to install apps.",
	},

	ErrUserNotFound: {
		Code:    ErrUserNotFound,
		Message: "User cannot be found",
	},

	ErrUserRemovedFromTeam: {
		Code:    ErrUserRemovedFromTeam,
		Message: "User removed from team (generated)",
	},

	ErrWorkflowNotFound: {
		Code:    ErrWorkflowNotFound,
		Message: "Workflow not found",
	},

	ErrYaml: {
		Code:    ErrYaml,
		Message: "An error occurred while parsing the app manifest YAML file",
	},
}
