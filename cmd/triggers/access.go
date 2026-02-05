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

package triggers

import (
	"context"
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type accessCmdFlags struct {
	users            string
	grant            bool
	revoke           bool
	everyone         bool
	appCollab        bool
	info             bool
	triggerID        string
	channels         string
	workspaces       string
	organizations    string
	includeAppCollab bool
}

var accessFlags accessCmdFlags

var accessAppSelectPromptFunc = prompts.AppSelectPrompt

func NewAccessCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access --trigger-id <id> [flags]",
		Short: "Manage who can use your triggers",
		Long:  "Manage who can use your triggers",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "trigger access --trigger-id Ft01234ABCD --everyone", Meaning: "Grant everyone access to run a trigger"},
			{Command: "trigger access --trigger-id Ft01234ABCD --grant \\\n    --channels C012345678", Meaning: "Grant certain channels access to run a trigger"},
			{Command: "trigger access --trigger-id Ft01234ABCD --revoke \\\n    --users USLACKBOT,U012345678", Meaning: "Revoke certain users access to run a trigger"},
		}),
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAccessCommand(cmd, clients)
		},
	}

	cmd.Flags().StringVarP(&accessFlags.triggerID, "trigger-id", "T", "", "the ID of the trigger")

	cmd.Flags().StringVarP(&accessFlags.users, "users", "U", "", "a comma-separated list of Slack user IDs")
	cmd.Flags().StringVarP(&accessFlags.channels, "channels", "C", "", "a comma-separated list of Slack channel IDs")
	cmd.Flags().StringVarP(&accessFlags.workspaces, "workspaces", "W", "", "a comma-separated list of Slack workspace IDs")
	cmd.Flags().StringVarP(&accessFlags.organizations, "organizations", "O", "", "a comma-separated list of Slack organization IDs")

	cmd.Flags().BoolVarP(&accessFlags.grant, "grant", "G", false, "grant permission to --users or --channels to\n  run the trigger --trigger-id")
	cmd.Flags().BoolVarP(&accessFlags.revoke, "revoke", "R", false, "revoke permission for --users or --channels to\n  run the trigger --trigger-id")

	cmd.Flags().BoolVarP(&accessFlags.everyone, "everyone", "E", false, "grant permission to everyone in your workspace")
	cmd.Flags().BoolVarP(&accessFlags.appCollab, "app-collaborators", "A", false, "grant permission to only app collaborators")
	cmd.Flags().BoolVarP(&accessFlags.info, "info", "I", false, "check who has access to the trigger --trigger-id")

	cmd.Flags().BoolVar(&accessFlags.includeAppCollab, "include-app-collaborators", false, "include app collaborators into named\n entities to run the trigger --trigger-id")

	return cmd
}

// runAccessCommand will execute the access command
func runAccessCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()
	var span, _ = opentracing.StartSpanFromContext(ctx, "cmd.triggers.access")
	defer span.Finish()

	// Get the app selection and accompanying auth from the flag or prompt
	selection, err := accessAppSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}
	token := selection.Auth.Token
	ctx = config.SetContextToken(ctx, token)
	app := selection.App

	if err = cmdutil.AppExists(app, selection.Auth); err != nil {
		return err
	}

	// Get trigger ID from flag or prompt
	if accessFlags.triggerID == "" {
		accessFlags.triggerID, err = promptForTriggerID(ctx, cmd, clients, app, token, labelsIncludeAccessType)
		if err != nil {
			if slackerror.ToSlackError(err).Code == slackerror.ErrNoTriggers {
				printNoTriggersMessage(ctx, clients.IO)
				return nil
			}
			return err
		}
	}

	// If --info flag is passed, execution ends here
	if accessFlags.info {
		return printAccess(cmd, clients, selection.Auth.Token, selection.App)
	}

	// Get the current access for the trigger
	currentAccessType, currentAuthorizedEntities, err := clients.API().TriggerPermissionsList(ctx, token, accessFlags.triggerID)
	if err != nil {
		return err
	}

	accessNamedEntities := nonEmptyNamedEntities()

	// Get new access type from flag or prompt
	var accessType types.Permission
	if !accessFlags.everyone && !accessFlags.appCollab && accessNamedEntities == 0 {
		if accessFlags.grant == accessFlags.revoke && accessFlags.grant {
			return slackerror.New(slackerror.ErrMismatchedFlags).WithMessage("Please specify '--grant' or '--revoke'.")
		}

		if accessFlags.revoke {
			if currentAccessType != types.PermissionNamedEntities {
				return slackerror.New("Trigger access permission is not set to specific entities, grant an entity access first")
			} else {
				accessType = types.PermissionNamedEntities
			}
		} else {
			accessType, err = promptForAccessType(ctx, clients, token, currentAccessType)
			if err != nil {
				return err
			}
		}
	} else {
		accessType = getPermissionTypeFromFlags()
	}

	// Set access type
	if accessType == types.PermissionEveryone || accessType == types.PermissionAppCollaborators {
		_, err = clients.API().TriggerPermissionsSet(ctx, token, accessFlags.triggerID, "", accessType, "")
		if err != nil {
			return err
		}
	}

	// Add or remove users, channels, workspaces or organizations from the named_entities ACL as per flags or prompts
	if accessType == types.PermissionNamedEntities {
		err := manageNamedEntities(cmd, clients, selection.Auth.Token, selection.App, currentAccessType, currentAuthorizedEntities)
		if err != nil {
			return err
		}
	}

	return printAccess(cmd, clients, selection.Auth.Token, selection.App)
}

func promptForAccessType(ctx context.Context, clients *shared.ClientFactory, token string, currentAccessType types.Permission) (types.Permission, error) {
	selectedPermission := new(types.Permission)
	accessOptionLabels, permissions := prompts.TriggerAccessLabels(currentAccessType)
	selection, err := clients.IO.SelectPrompt(ctx, "Who can find and run this trigger?", accessOptionLabels, iostreams.SelectPromptConfig{
		Flags: []*pflag.Flag{
			clients.Config.Flags.Lookup("channels"),
			clients.Config.Flags.Lookup("organizations"),
			clients.Config.Flags.Lookup("users"),
			clients.Config.Flags.Lookup("workspaces"),
			clients.Config.Flags.Lookup("app-collaborators"),
			clients.Config.Flags.Lookup("everyone"),
		},
		Required: true,
	})
	if err != nil {
		if slackerror.ToSlackError(err).Code == slackerror.ErrMismatchedFlags &&
			types.IsNamedEntityFlag(clients.Config.Flags) {
			return types.PermissionNamedEntities, nil
		}
		return *selectedPermission, err
	} else if selection.Flag && types.IsNamedEntityFlag(clients.Config.Flags) {
		return types.PermissionNamedEntities, nil
	} else if selection.Flag && clients.Config.Flags.Lookup("app-collaborators").Changed {
		return types.PermissionAppCollaborators, nil
	} else if selection.Flag && clients.Config.Flags.Lookup("everyone").Changed {
		return types.PermissionEveryone, nil
	} else if selection.Prompt {
		return permissions[selection.Index], nil
	}
	return *selectedPermission, nil
}

// getPermissionTypeFromFlags returns the permission type from the flags
func getPermissionTypeFromFlags() types.Permission {
	if accessFlags.appCollab {
		return types.PermissionAppCollaborators
	}

	if accessFlags.users != "" || accessFlags.channels != "" || accessFlags.workspaces != "" || accessFlags.organizations != "" {
		return types.PermissionNamedEntities
	}

	return types.PermissionEveryone
}

func manageNamedEntities(cmd *cobra.Command, clients *shared.ClientFactory, token string, app types.App, currentAccessType types.Permission, currentAuthorizedEntities []string) error {
	ctx := cmd.Context()

	accessFlags.users = goutils.UpperCaseTrimAll(accessFlags.users)
	accessFlags.channels = goutils.UpperCaseTrimAll(accessFlags.channels)
	accessFlags.workspaces = goutils.UpperCaseTrimAll(accessFlags.workspaces)
	accessFlags.organizations = goutils.UpperCaseTrimAll(accessFlags.organizations)
	accessNamedEntities := nonEmptyNamedEntities()
	action := ""
	includeAppCollaborators := false

	// set includeAppCollaborators only when `include-app-collaborators` flag is called in command
	// If the flag is provided, we skip the prompt and set includeAppCollaborators as the flag value
	// If the flag is not provided, display the prompt to include app collaborators
	if cmdutil.IsFlagChanged(cmd, "include-app-collaborators") {
		includeAppCollaborators = accessFlags.includeAppCollab
	}

	// prompt if list of named_entities not passed in, and one of 'grant' or 'revoke' is not specified
	if accessNamedEntities == 0 || accessFlags.grant == accessFlags.revoke {
		namedEntities := ""
		accessAction, err := prompts.TriggerChooseNamedEntityActionPrompt(ctx, clients)
		if err != nil {
			return err
		}
		if accessAction == "cancel" {
			return nil
		}

		if accessNamedEntities > 0 {
			switch accessAction {
			case "grant":
				accessFlags.grant = true
				if !cmdutil.IsFlagChanged(cmd, "include-app-collaborators") && currentAccessType != types.PermissionNamedEntities {
					includeAppCollaborators, err = prompts.AddAppCollaboratorsToNamedEntitiesPrompt(ctx, clients.IO)
					if err != nil {
						return err
					}
				}
			case "revoke":
				accessFlags.revoke = true
			default:
				return nil
			}
		} else {
			err := printCurrentAuthorizedEntities(cmd, clients, token, app, currentAuthorizedEntities, currentAccessType)
			if err != nil {
				return err
			}

			action, namedEntities, includeAppCollaborators, err = prompts.TriggerChooseNamedEntityPrompt(ctx, clients, accessAction, currentAccessType, cmdutil.IsFlagChanged(cmd, "include-app-collaborators"))
			// Overwrite includeAppCollaborators from TriggerChooseNamedEntityPrompt() if flag is set
			if cmdutil.IsFlagChanged(cmd, "include-app-collaborators") {
				includeAppCollaborators = accessFlags.includeAppCollab
			}

			if err != nil {
				return err
			}
			if action == "cancel" {
				return nil
			}

			if strings.Contains(action, "_user") {
				accessFlags.users = namedEntities
			} else if strings.Contains(action, "_channel") {
				accessFlags.channels = namedEntities
			} else if strings.Contains(action, "_workspace") {
				accessFlags.workspaces = namedEntities
			} else if strings.Contains(action, "_organization") {
				accessFlags.organizations = namedEntities
			}
		}
	} else {
		if !cmdutil.IsFlagChanged(cmd, "include-app-collaborators") && currentAccessType != types.PermissionNamedEntities && accessFlags.grant {
			var err error
			includeAppCollaborators, err = prompts.AddAppCollaboratorsToNamedEntitiesPrompt(ctx, clients.IO)
			if err != nil {
				return err
			}
		}
	}

	if includeAppCollaborators && currentAccessType != types.PermissionNamedEntities {
		err := AddAppCollaboratorsToNamedEntities(ctx, clients, token, app.AppID)
		if err != nil {
			return err
		}
	}

	if accessFlags.revoke {
		if accessNamedEntities > 1 {
			action = "remove_entities"
		} else {
			if accessFlags.users != "" {
				action = "remove_user"
			} else if accessFlags.channels != "" {
				action = "remove_channel"
			} else if accessFlags.workspaces != "" {
				action = "remove_workspace"
			} else if accessFlags.organizations != "" {
				action = "remove_organization"
			}
		}
	}
	if accessFlags.grant {
		if accessNamedEntities > 1 {
			action = "add_entities"
		} else {
			if accessFlags.users != "" {
				action = "add_user"
			} else if accessFlags.channels != "" {
				action = "add_channel"
			} else if accessFlags.workspaces != "" {
				action = "add_workspace"
			} else if accessFlags.organizations != "" {
				action = "add_organization"
			}
		}
	}

	switch action {
	case "add_user":
		if currentAccessType != types.PermissionNamedEntities {
			if includeAppCollaborators {
				err := clients.API().TriggerPermissionsAddEntities(ctx, token, accessFlags.triggerID, accessFlags.users, "users")
				if err != nil {
					return err
				}
			} else {
				_, err := clients.API().TriggerPermissionsSet(ctx, token, accessFlags.triggerID, accessFlags.users, types.PermissionNamedEntities, "users")
				if err != nil {
					return err
				}
			}
		} else {
			err := clients.API().TriggerPermissionsAddEntities(ctx, token, accessFlags.triggerID, accessFlags.users, "users")
			if err != nil {
				return err
			}
		}

		users := strings.Split(accessFlags.users, ",")
		clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s added %s", style.Pluralize("User", "Users", len(users)), style.Emoji("party_popper"))))

	case "remove_user":
		if currentAccessType == types.PermissionNamedEntities && len(currentAuthorizedEntities) == 0 {
			return slackerror.New("Access list is empty; cannot remove from an empty list")
		}
		if currentAccessType != types.PermissionNamedEntities {
			return slackerror.New("Grant a user access first")
		}

		err := clients.API().TriggerPermissionsRemoveEntities(ctx, token, accessFlags.triggerID, accessFlags.users, "users")
		if err != nil {
			return err
		}

		users := strings.Split(accessFlags.users, ",")
		clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s removed %s", style.Pluralize("User", "Users", len(users)), style.Emoji("firecracker"))))

	case "add_channel":
		if currentAccessType != types.PermissionNamedEntities {
			if includeAppCollaborators {
				err := clients.API().TriggerPermissionsAddEntities(ctx, token, accessFlags.triggerID, accessFlags.channels, "channels")
				if err != nil {
					return err
				}
			} else {
				_, err := clients.API().TriggerPermissionsSet(ctx, token, accessFlags.triggerID, accessFlags.channels, types.PermissionNamedEntities, "channels")
				if err != nil {
					return err
				}
			}
		} else {
			err := clients.API().TriggerPermissionsAddEntities(ctx, token, accessFlags.triggerID, accessFlags.channels, "channels")
			if err != nil {
				return err
			}
		}

		channels := strings.Split(accessFlags.channels, ",")
		clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s added %s", style.Pluralize("Channel", "Channels", len(channels)), style.Emoji("party_popper"))))

	case "remove_channel":
		if currentAccessType == types.PermissionNamedEntities && len(currentAuthorizedEntities) == 0 {
			return slackerror.New("Access list is empty; cannot remove from an empty list")
		}
		if currentAccessType != types.PermissionNamedEntities {
			return slackerror.New("Grant a channel access first")
		}

		err := clients.API().TriggerPermissionsRemoveEntities(ctx, token, accessFlags.triggerID, accessFlags.channels, "channels")
		if err != nil {
			return err
		}

		channels := strings.Split(accessFlags.channels, ",")
		clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s removed %s", style.Pluralize("Channel", "Channels", len(channels)), style.Emoji("firecracker"))))

	case "add_workspace":
		if currentAccessType != types.PermissionNamedEntities {
			if includeAppCollaborators {
				err := clients.API().TriggerPermissionsAddEntities(ctx, token, accessFlags.triggerID, accessFlags.workspaces, "workspaces")
				if err != nil {
					return err
				}
			} else {
				_, err := clients.API().TriggerPermissionsSet(ctx, token, accessFlags.triggerID, accessFlags.workspaces, types.PermissionNamedEntities, "workspaces")
				if err != nil {
					return err
				}
			}
		} else {
			err := clients.API().TriggerPermissionsAddEntities(ctx, token, accessFlags.triggerID, accessFlags.workspaces, "workspaces")
			if err != nil {
				return err
			}
		}

		workspaces := strings.Split(accessFlags.workspaces, ",")
		clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s added %s", style.Pluralize("Workspace", "Workspaces", len(workspaces)), style.Emoji("party_popper"))))

	case "remove_workspace":
		if currentAccessType == types.PermissionNamedEntities && len(currentAuthorizedEntities) == 0 {
			return slackerror.New("Access list is empty; cannot remove from an empty list")
		}
		if currentAccessType != types.PermissionNamedEntities {
			return slackerror.New("Grant a workspace access first")
		}

		err := clients.API().TriggerPermissionsRemoveEntities(ctx, token, accessFlags.triggerID, accessFlags.workspaces, "workspaces")
		if err != nil {
			return err
		}

		workspaces := strings.Split(accessFlags.workspaces, ",")
		clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s removed %s", style.Pluralize("Workspace", "Workspaces", len(workspaces)), style.Emoji("firecracker"))))

	case "add_organization":
		if currentAccessType != types.PermissionNamedEntities {
			if includeAppCollaborators {
				err := clients.API().TriggerPermissionsAddEntities(ctx, token, accessFlags.triggerID, accessFlags.organizations, "organizations")
				if err != nil {
					return err
				}
			} else {
				_, err := clients.API().TriggerPermissionsSet(ctx, token, accessFlags.triggerID, accessFlags.organizations, types.PermissionNamedEntities, "organizations")
				if err != nil {
					return err
				}
			}
		} else {
			err := clients.API().TriggerPermissionsAddEntities(ctx, token, accessFlags.triggerID, accessFlags.organizations, "organizations")
			if err != nil {
				return err
			}
		}

		organizations := strings.Split(accessFlags.organizations, ",")
		clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s added %s", style.Pluralize("Organization", "Organizations", len(organizations)), style.Emoji("party_popper"))))

	case "remove_organization":
		if currentAccessType == types.PermissionNamedEntities && len(currentAuthorizedEntities) == 0 {
			return slackerror.New("Access list is empty; cannot remove from an empty list")
		}
		if currentAccessType != types.PermissionNamedEntities {
			return slackerror.New("Grant an organization access first")
		}

		err := clients.API().TriggerPermissionsRemoveEntities(ctx, token, accessFlags.triggerID, accessFlags.organizations, "organizations")
		if err != nil {
			return err
		}

		organizations := strings.Split(accessFlags.organizations, ",")
		clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s removed %s", style.Pluralize("Organization", "Organizations", len(organizations)), style.Emoji("firecracker"))))

	case "add_entities":
		if currentAccessType != types.PermissionNamedEntities {
			index := 0
			for namedEntityType, namedEntityVal := range namedEntitiesValMap() {
				if index == 0 && !includeAppCollaborators {
					_, triggerSetErr := clients.API().TriggerPermissionsSet(ctx, token, accessFlags.triggerID, namedEntityVal, types.PermissionNamedEntities, namedEntityType)
					if triggerSetErr != nil {
						return triggerSetErr
					}
				} else {
					err := clients.API().TriggerPermissionsAddEntities(ctx, token, accessFlags.triggerID, namedEntityVal, namedEntityType)
					if err != nil {
						return err
					}
				}
				index++
				namedEntityValList := strings.Split(namedEntityVal, ",")
				clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s added %s", style.Pluralize(cases.Title(language.Und, cases.NoLower).String(strings.TrimSuffix(namedEntityType, "s")), cases.Title(language.Und, cases.NoLower).String(namedEntityType), len(namedEntityValList)), style.Emoji("party_popper"))))
			}
		} else {
			for namedEntityType, namedEntityVal := range namedEntitiesValMap() {
				err := clients.API().TriggerPermissionsAddEntities(ctx, token, accessFlags.triggerID, namedEntityVal, namedEntityType)
				if err != nil {
					return err
				}
				namedEntityValList := strings.Split(namedEntityVal, ",")
				clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s added %s", style.Pluralize(cases.Title(language.Und, cases.NoLower).String(strings.TrimSuffix(namedEntityType, "s")), cases.Title(language.Und, cases.NoLower).String(namedEntityType), len(namedEntityValList)), style.Emoji("party_popper"))))
			}
		}

	case "remove_entities":
		if currentAccessType == types.PermissionNamedEntities && len(currentAuthorizedEntities) == 0 {
			return slackerror.New("Access list is empty; cannot remove from an empty list")
		}
		if currentAccessType != types.PermissionNamedEntities {
			return slackerror.New("Grant an entity access first")
		}

		for namedEntityType, namedEntityVal := range namedEntitiesValMap() {
			err := clients.API().TriggerPermissionsRemoveEntities(ctx, token, accessFlags.triggerID, namedEntityVal, namedEntityType)
			if err != nil {
				return err
			}
			namedEntityValList := strings.Split(namedEntityVal, ",")
			clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s removed %s", style.Pluralize(cases.Title(language.Und, cases.NoLower).String(strings.TrimSuffix(namedEntityType, "s")), cases.Title(language.Und, cases.NoLower).String(namedEntityType), len(namedEntityValList)), style.Emoji("firecracker"))))
		}
	}
	return nil
}

// printAccess formats and displays access information
func printAccess(cmd *cobra.Command, clients *shared.ClientFactory, token string, app types.App) error {
	ctx := cmd.Context()

	accessType, userAccessList, err := clients.API().TriggerPermissionsList(ctx, token, accessFlags.triggerID)
	if err != nil {
		clients.IO.PrintTrace(ctx, slacktrace.TriggersAccessError)
		return err
	}

	switch accessType {
	case types.PermissionEveryone:
		var everyoneAccessTypeDescription = types.GetAccessTypeDescriptionForEveryone(app)
		clients.IO.PrintInfo(ctx, false, "\nTrigger '%s' can be found and run by %s", accessFlags.triggerID, everyoneAccessTypeDescription)
	case types.PermissionAppCollaborators:
		clients.IO.PrintInfo(ctx, false, "\nTrigger '%s' can be found and run by %s:", accessFlags.triggerID, style.Pluralize("app collaborator", "app collaborators", len(userAccessList)))
		err = printAppCollaboratorsHelper(cmd, clients, token, userAccessList)
	case types.PermissionNamedEntities:
		err = printNamedEntitiesHelper(cmd, clients, token, userAccessList, "list")
	}
	clients.IO.PrintTrace(ctx, slacktrace.TriggersAccessSuccess)
	return err
}

// printCurrentAuthorizedEntities formats and displays current access information
func printCurrentAuthorizedEntities(cmd *cobra.Command, clients *shared.ClientFactory, token string, app types.App, currentAccessList []string, currentAccessType types.Permission) error {
	ctx := cmd.Context()

	cmd.Println()
	switch currentAccessType {
	case types.PermissionEveryone:
		var everyoneAccessTypeDescription = types.GetAccessTypeDescriptionForEveryone(app)
		clients.IO.PrintInfo(ctx, false, "Trigger '%s' can be found and run by %s\n", accessFlags.triggerID, everyoneAccessTypeDescription)
	case types.PermissionAppCollaborators:
		clients.IO.PrintInfo(ctx, false, "Access is currently granted to %s:", style.Pluralize("app collaborator", "app collaborators", len(currentAccessList)))
		err := printAppCollaboratorsHelper(cmd, clients, token, currentAccessList)
		if err != nil {
			return err
		}
	case types.PermissionNamedEntities:
		err := printNamedEntitiesHelper(cmd, clients, token, currentAccessList, "manage")
		if err != nil {
			return err
		}
	}
	return nil
}

func printAppCollaboratorsHelper(cmd *cobra.Command, clients *shared.ClientFactory, token string, userAccessList []string) error {
	ctx := cmd.Context()

	if len(userAccessList) <= 0 {
		clients.IO.PrintInfo(ctx, false, "nobody")
		return nil
	}

	for _, entity := range userAccessList {
		userInfo, err := clients.API().UsersInfo(ctx, token, entity)
		if err != nil {
			return err
		}
		clients.IO.PrintInfo(ctx, false, "  %s %s %s", userInfo.RealName, style.Secondary("@"+userInfo.Profile.DisplayName), style.Secondary(userInfo.ID))
	}
	cmd.Println()
	return nil
}

func printNamedEntitiesHelper(cmd *cobra.Command, clients *shared.ClientFactory, token string, entitiesAccessList []string, action string) error {
	ctx := cmd.Context()

	if len(entitiesAccessList) <= 0 {
		switch action {
		case "manage":
			clients.IO.PrintInfo(ctx, false, "Access is currently granted:")
		case "list":
			clients.IO.PrintInfo(ctx, false, "\nTrigger '%s' can be found and run by:", accessFlags.triggerID)
		}
		clients.IO.PrintInfo(ctx, false, "nobody")
		return nil
	}

	namedEntitiesAccessMap := namedEntitiesAccessMap(entitiesAccessList)

	if len(namedEntitiesAccessMap["users"]) > 0 {
		var userLabel = style.Pluralize("this user", "these users", len(namedEntitiesAccessMap["users"]))
		switch action {
		case "manage":
			clients.IO.PrintInfo(ctx, false, "\nAccess is currently granted to %s:", userLabel)
		case "list":
			clients.IO.PrintInfo(ctx, false, "\nTrigger '%s' can be found and run by %s:", accessFlags.triggerID, userLabel)
		}
		for _, entity := range namedEntitiesAccessMap["users"] {
			userInfo, err := clients.API().UsersInfo(ctx, token, entity)
			if err != nil {
				return err
			}
			clients.IO.PrintInfo(ctx, false, "  %s %s %s", userInfo.RealName, style.Secondary("@"+userInfo.Profile.DisplayName), style.Secondary(userInfo.ID))
		}
	}
	if len(namedEntitiesAccessMap["channels"]) > 0 {
		var channelLabel = style.Pluralize("this channel", "these channels", len(namedEntitiesAccessMap["channels"]))
		switch action {
		case "manage":
			clients.IO.PrintInfo(ctx, false, "\nAccess is currently granted to all members of %s:", channelLabel)
		case "list":
			clients.IO.PrintInfo(ctx, false, "\nTrigger '%s' can be found and run by all members of %s:", accessFlags.triggerID, channelLabel)
		}
		for _, entity := range namedEntitiesAccessMap["channels"] {
			channelInfo, err := clients.API().ChannelsInfo(ctx, token, entity)
			if err != nil {
				return err
			}
			clients.IO.PrintInfo(ctx, false, "  #%s %s", channelInfo.Name, style.Secondary(channelInfo.ID))
		}
	}
	if len(namedEntitiesAccessMap["teams"]) > 0 {
		var teamLabel = style.Pluralize("this workspace", "these workspaces", len(namedEntitiesAccessMap["teams"]))
		switch action {
		case "manage":
			clients.IO.PrintInfo(ctx, false, "\nAccess is currently granted to all members of %s:", teamLabel)
		case "list":
			clients.IO.PrintInfo(ctx, false, "\nTrigger '%s' can be found and run by all members of %s:", accessFlags.triggerID, teamLabel)
		}
		for _, entity := range namedEntitiesAccessMap["teams"] {
			teamInfo, err := clients.API().TeamsInfo(ctx, token, entity)
			if err != nil {
				return err
			}
			clients.IO.PrintInfo(ctx, false, "  %s %s", teamInfo.Name, style.Secondary(teamInfo.ID))
		}
	}
	if len(namedEntitiesAccessMap["organizations"]) > 0 {
		var orgLabel = style.Pluralize("this organization", "these organizations", len(namedEntitiesAccessMap["organizations"]))
		switch action {
		case "manage":
			clients.IO.PrintInfo(ctx, false, "\nAccess is currently granted to all members of %s:", orgLabel)
		case "list":
			clients.IO.PrintInfo(ctx, false, "\nTrigger '%s' can be found and run by all members of %s:", accessFlags.triggerID, orgLabel)
		}
		for _, entity := range namedEntitiesAccessMap["organizations"] {
			orgInfo, err := clients.API().TeamsInfo(ctx, token, entity)
			if err != nil {
				return err
			}
			clients.IO.PrintInfo(ctx, false, "  %s %s", orgInfo.Name, style.Secondary(orgInfo.ID))
		}
	}
	cmd.Println()
	return nil
}

// nonEmptyNamedEntities returns number of passed named_entities types from user
func nonEmptyNamedEntities() int {
	givenNamedEntities := 0
	namedEntities := []string{accessFlags.users, accessFlags.channels, accessFlags.workspaces, accessFlags.organizations}
	for _, v := range namedEntities {
		if goutils.UpperCaseTrimAll(v) != "" {
			givenNamedEntities++
		}
	}
	return givenNamedEntities
}

// namedEntitiesValMap returns a map with key as named_entities type and value as what user passes in flag
func namedEntitiesValMap() map[string]string {
	namedEntitiesMap := make(map[string]string)
	namedEntitiesMap["users"] = goutils.UpperCaseTrimAll(accessFlags.users)
	namedEntitiesMap["channels"] = goutils.UpperCaseTrimAll(accessFlags.channels)
	namedEntitiesMap["workspaces"] = goutils.UpperCaseTrimAll(accessFlags.workspaces)
	namedEntitiesMap["organizations"] = goutils.UpperCaseTrimAll(accessFlags.organizations)

	for k, v := range namedEntitiesMap {
		if v == "" {
			delete(namedEntitiesMap, k)
		}
	}

	return namedEntitiesMap
}

// namedEntitiesAccessMap returns a map with key as named_entities type and value as slice of named_entities value
func namedEntitiesAccessMap(entitiesAccessList []string) map[string][]string {
	namedEntitiesAccessMap := make(map[string][]string)

	for _, entity := range entitiesAccessList {
		if strings.HasPrefix(entity, "U") {
			namedEntitiesAccessMap["users"] = append(namedEntitiesAccessMap["users"], entity)
		} else if strings.HasPrefix(entity, "C") {
			namedEntitiesAccessMap["channels"] = append(namedEntitiesAccessMap["channels"], entity)
		} else if strings.HasPrefix(entity, "T") {
			namedEntitiesAccessMap["teams"] = append(namedEntitiesAccessMap["teams"], entity)
		} else if strings.HasPrefix(entity, "E") {
			namedEntitiesAccessMap["organizations"] = append(namedEntitiesAccessMap["organizations"], entity)
		}
	}
	return namedEntitiesAccessMap
}

// AddAppCollaboratorsToNamedEntities adds app_collaborators to named_entities list if trigger ACL is changed to named_entities
func AddAppCollaboratorsToNamedEntities(ctx context.Context, clients *shared.ClientFactory, token string, appID string) error {
	ctx = config.SetContextToken(ctx, token)

	collaborators, err := clients.API().ListCollaborators(ctx, token, appID)
	if err != nil {
		return err
	}

	if len(collaborators) == 0 {
		return nil
	}

	userIDs := collaborators[0].ID
	for i := 1; i < len(collaborators); i++ {
		userIDs = userIDs + "," + collaborators[i].ID
	}

	_, err = clients.API().TriggerPermissionsSet(ctx, token, accessFlags.triggerID, userIDs, types.PermissionNamedEntities, "users")
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, style.Secondary(fmt.Sprintf("%s added %s", style.Pluralize("App collaborator", "App collaborators", len(collaborators)), style.Emoji("party_popper"))))
	return nil
}
