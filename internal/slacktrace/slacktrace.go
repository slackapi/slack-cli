// Copyright 2022-2026 Salesforce, Inc.
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

package slacktrace

// Tracer IDs were introduced here:
// https://github.com/slackapi/slack-cli/pull/846
//
// Usage:
//
// Set the environment variable SLACK_TEST_TRACE=true
//
// - Run a single command:
//   - SLACK_TEST_TRACE=true slack <command>
//
// Naming convention:
//
// - Id Format:
//   - Key:   CommandName<Status>          (e.g. CreateDependenciesSuccess)
//   - Value: SLACK_TRACE_<COMMAND_STATUS> (e.g. SLACK_TRACE_CREATE_DEPENDENCIES_SUCCESS)
//
// - Order: Alphabetical
//
// To add trace output to a command:
//
// use,
// clients.IO.PrintTrace(ctx, slacktrace.<NAMEDCONST>)
const (
	AdminAppApprovalRequestPending         = "SLACK_TRACE_ADMIN_APPROVAL_REQUEST_PENDING"
	AdminAppApprovalRequestReasonSubmitted = "SLACK_TRACE_ADMIN_APPROVAL_REQUEST_REASON_SUBMITTED"
	AdminAppApprovalRequestSendError       = "SLACK_TRACE_ADMIN_APPROVAL_REQUEST_SEND_ERROR"
	AdminAppApprovalRequestShouldSend      = "SLACK_TRACE_ADMIN_APPROVAL_REQUEST_SHOULD_SEND"
	AdminAppApprovalRequestRequired        = "SLACK_TRACE_ADMIN_APPROVAL_REQUIRED"
	AppLinkStart                           = "SLACK_TRACE_APP_LINK_START"
	AppLinkSuccess                         = "SLACK_TRACE_APP_LINK_SUCCESS"
	AppSettingsStart                       = "SLACK_TRACE_APP_SETTINGS_START"
	AppSettingsSuccess                     = "SLACK_TRACE_APP_SETTINGS_SUCCESS"
	AppUnlinkStart                         = "SLACK_TRACE_APP_UNLINK_START"
	AppUnlinkSuccess                       = "SLACK_TRACE_APP_UNLINK_SUCCESS"
	AuthListCount                          = "SLACK_TRACE_AUTH_LIST_COUNT"
	AuthListInfo                           = "SLACK_TRACE_AUTH_LIST_INFO"
	AuthListSuccess                        = "SLACK_TRACE_AUTH_LIST_SUCCESS"
	AuthLoginStart                         = "SLACK_TRACE_AUTH_LOGIN_START"
	AuthLoginSuccess                       = "SLACK_TRACE_AUTH_LOGIN_SUCCESS"
	AuthLogoutStart                        = "SLACK_TRACE_AUTH_LOGOUT_START"
	AuthLogoutSuccess                      = "SLACK_TRACE_AUTH_LOGOUT_SUCCESS"
	AuthRevokeStart                        = "SLACK_TRACE_AUTH_REVOKE_START"
	AuthRevokeSuccess                      = "SLACK_TRACE_AUTH_REVOKE_SUCCESS"
	CollaboratorAddCollaborator            = "SLACK_TRACE_COLLABORATOR_ADD_COLLABORATOR"
	CollaboratorAddSuccess                 = "SLACK_TRACE_COLLABORATOR_ADD_SUCCESS"
	CollaboratorListCollaborator           = "SLACK_TRACE_COLLABORATOR_LIST_COLLABORATOR"
	CollaboratorListCount                  = "SLACK_TRACE_COLLABORATOR_LIST_COUNT"
	CollaboratorListSuccess                = "SLACK_TRACE_COLLABORATOR_LIST_SUCCESS"
	CollaboratorRemoveCollaborator         = "SLACK_TRACE_COLLABORATOR_REMOVE_COLLABORATOR"
	CollaboratorRemoveSuccess              = "SLACK_TRACE_COLLABORATOR_REMOVE_SUCCESS"
	CreateCategoryOptions                  = "SLACK_TRACE_CREATE_CATEGORY_OPTIONS"
	CreateDependenciesSuccess              = "SLACK_TRACE_CREATE_DEPENDENCIES_SUCCESS"
	CreateError                            = "SLACK_TRACE_CREATE_ERROR"
	CreateProjectPath                      = "SLACK_TRACE_CREATE_PROJECT_PATH"
	CreateStart                            = "SLACK_TRACE_CREATE_START"
	CreateSuccess                          = "SLACK_TRACE_CREATE_SUCCESS"
	CreateTemplateOptions                  = "SLACK_TRACE_CREATE_TEMPLATE_OPTIONS"
	DatastoreCountDatastore                = "SLACK_TRACE_DATASTORE_COUNT_DATASTORE"
	DatastoreCountSuccess                  = "SLACK_TRACE_DATASTORE_COUNT_SUCCESS"
	DatastoreCountTotal                    = "SLACK_TRACE_DATASTORE_COUNT_TOTAL"
	EnvAddSuccess                          = "SLACK_TRACE_ENV_ADD_SUCCESS"
	EnvListCount                           = "SLACK_TRACE_ENV_LIST_COUNT"
	EnvListVariables                       = "SLACK_TRACE_ENV_LIST_VARIABLES"
	EnvRemoveSuccess                       = "SLACK_TRACE_ENV_REMOVE_SUCCESS"
	FeedbackMessage                        = "SLACK_TRACE_FEEDBACK_MESSAGE"
	ManifestValidateSuccess                = "SLACK_TRACE_MANIFEST_VALIDATE_SUCCESS"
	PlatformDeploySuccess                  = "SLACK_TRACE_PLATFORM_DEPLOY_SUCCESS"
	PlatformRunReady                       = "SLACK_TRACE_PLATFORM_RUN_READY"
	PlatformRunStart                       = "SLACK_TRACE_PLATFORM_RUN_START"
	PlatformRunStop                        = "SLACK_TRACE_PLATFORM_RUN_STOP"
	ProjectInitStarted                     = "SLACK_TRACE_PROJECT_INIT_STARTED"
	ProjectInitSuccess                     = "SLACK_TRACE_PROJECT_INIT_SUCCESS"
	TriggersAccessError                    = "SLACK_TRACE_TRIGGERS_ACCESS_ERROR"
	TriggersAccessSuccess                  = "SLACK_TRACE_TRIGGERS_ACCESS_SUCCESS"
	TriggersCreateSuccess                  = "SLACK_TRACE_TRIGGERS_CREATE_SUCCESS"
	TriggersCreateURL                      = "SLACK_TRACE_TRIGGERS_CREATE_URL"
)
