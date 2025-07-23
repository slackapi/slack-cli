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

package function

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type distributeFlagSet struct {
	users     string
	file      string
	grant     bool
	revoke    bool
	everyone  bool
	appCollab bool
	info      bool
}

var distributeFlags distributeFlagSet

func NewDistributeCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access",
		Short: "Adjust who can access functions published from an app",
		Long: strings.Join([]string{
			`Adjust who can {{ToBold "access"}} functions published by an app when building a workflow in`,
			"Workflow Builder.",
			"",
			`New functions are granted access to {{ToBold "app collaborators"}} by default. This includes`,
			`both the {{ToBold "reader"}} and {{ToBold "owner"}} permissions. Access can also be {{ToBold "granted"}} or {{ToBold "revoked"}} to`,
			`specific {{ToBold "users"}} or {{ToBold "everyone"}} alongside the {{ToBold "app collaborators"}}.`,
			"",
			"Workflows that include a function with limited access can still be invoked with",
			`a trigger of the workflow. The {{ToBold "access"}} command applies to Workflow Builder access`,
			"only.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "function access", Meaning: "Select a function and choose access options"},
			{Command: "function access --name callback_id --everyone", Meaning: "Share a function with everyone in a workspace"},
			{Command: "function access --name callback_id --revoke \\\n    --users USLACKBOT,U012345678,U0RHJTSPQ3", Meaning: "Revoke function access for multiple users"},
			{Command: "function access --info", Meaning: "Lookup access information for a function"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return runDistributeCommand(cmd, clients)
		},
		Aliases: []string{"distribute", "distribution", "dist"},
	}

	cmd.PersistentFlags().BoolVarP(&distributeFlags.appCollab, "app-collaborators", "A", false, "grant access to only fellow app collaborators")
	cmd.PersistentFlags().BoolVarP(&distributeFlags.everyone, "everyone", "E", false, "grant access to everyone in installed workspaces")
	cmd.PersistentFlags().StringVarP(&distributeFlags.file, "file", "F", "", "specify access permissions using a file")
	cmd.PersistentFlags().BoolVarP(&distributeFlags.grant, "grant", "G", false, "grant access to --users to use --name")
	cmd.PersistentFlags().BoolVarP(&distributeFlags.info, "info", "I", false, "check who has access to the function --name")
	// TODO: The flag name, variable name, and description all use different terms: --name flag for a callback_id that maps to a functionFlag variable. Consider supporting `--callback-id in for the next semver MAJOR
	cmd.PersistentFlags().StringVarP(&functionFlag, "name", "N", "", "the callback_id of a function in your app")
	cmd.PersistentFlags().BoolVarP(&distributeFlags.revoke, "revoke", "R", false, "revoke access for --users to use --name")
	cmd.PersistentFlags().StringVarP(&distributeFlags.users, "users", "U", "", "a comma-separated list of Slack user IDs")

	return cmd
}

// runDistributeCommand will execute the distribute command
func runDistributeCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	// Get the app selection and accompanying auth from the prompt
	selectedApp, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	// FIXME: Stop relying on context getters and setters
	token := selectedApp.Auth.Token
	ctx = config.SetContextToken(ctx, token)
	app := selectedApp.App

	// FIXME: Rely on prompts to detect flags for the functionCallbackID
	if functionFlag == "" && distributeFlags.file == "" {
		functions, err := GetAppFunctions(ctx, clients, app, selectedApp.Auth)
		if err != nil {
			return err
		}
		functionCallbackID, err := chooseFunctionPrompt(ctx, clients, functions)
		if err != nil {
			return err
		}
		functionFlag = functionCallbackID
	}

	if distributeFlags.info {
		return printDistribution(ctx, cmd, clients, app)
	}
	return handleUpdate(ctx, cmd, clients, app, token)
}

// handleUpdate checks the flags passed in and sets or updates access permissions accordingly
func handleUpdate(
	ctx context.Context,
	cmd *cobra.Command,
	clients *shared.ClientFactory,
	app types.App,
	token string,
) error {
	distributeFlags.users = strings.ToUpper(distributeFlags.users)
	distributeFlags.users = strings.ReplaceAll(distributeFlags.users, " ", "") // trim whitespace

	if distributeFlags.file != "" {
		return distributePermissionFile(ctx, clients, app, distributeFlags.file)
	} else if distributeFlags.everyone || distributeFlags.appCollab {
		distribution := types.PermissionEveryone
		if distributeFlags.appCollab {
			distribution = types.PermissionAppCollaborators
		}

		_, err := clients.API().FunctionDistributionSet(ctx, functionFlag, app.AppID, distribution, "")
		if err != nil {
			return err
		}
		return printDistribution(ctx, cmd, clients, app)
	} else if distributeFlags.grant {
		if distributeFlags.users == "" {
			return slackerror.New(slackerror.ErrMissingFlag).WithRemediation("To grant a user access, pass in their user ID with the --users flag: `--users U12345`")
		}

		err := handleDistributionType(ctx, clients, app)
		if err != nil {
			return err
		}

		err = clients.API().FunctionDistributionAddUsers(ctx, functionFlag, app.AppID, distributeFlags.users)
		if err != nil {
			return err
		}

		users := strings.Split(distributeFlags.users, ",")
		_, _ = clients.IO.WriteOut().Write([]byte(style.Sectionf(style.TextSection{
			Emoji: "party_popper",
			Text:  fmt.Sprintf("Function access granted to the provided %s", style.Pluralize("user", "users", len(users))),
		})))
		return printDistribution(ctx, cmd, clients, app)
	} else if distributeFlags.revoke {
		if distributeFlags.users == "" {
			return slackerror.New(slackerror.ErrMissingFlag).WithRemediation("To revoke a user's access, pass in their user ID with the --users flag: `--users U12345`")
		}

		err := clients.API().FunctionDistributionRemoveUsers(ctx, functionFlag, app.AppID, distributeFlags.users)
		if err != nil {
			return err
		}

		users := strings.Split(distributeFlags.users, ",")
		_, _ = clients.IO.WriteOut().Write([]byte(style.Sectionf(style.TextSection{
			Emoji: "firecracker",
			Text:  fmt.Sprintf("Function access revoked for the provided %s", style.Pluralize("user", "users", len(users))),
		})))
		return printDistribution(ctx, cmd, clients, app)
	} else {
		dist, err := chooseDistributionPrompt(ctx, clients, app, token)
		if err != nil {
			return err
		}

		if dist == types.PermissionNamedEntities {
			err := printEntityAccess(ctx, cmd, clients, app)
			if err != nil {
				return err
			}

			action, users, err := prompts.ChooseNamedEntityPrompt(ctx, clients)
			if err != nil {
				return err
			}

			switch action {
			case "add":
				err := handleDistributionType(ctx, clients, app)
				if err != nil {
					return err
				}

				err = clients.API().FunctionDistributionAddUsers(ctx, functionFlag, app.AppID, users)
				if err != nil {
					return err
				}

				users := strings.Split(distributeFlags.users, ",")
				_, _ = clients.IO.WriteOut().Write([]byte(style.Sectionf(style.TextSection{
					Emoji: "party_popper",
					Text:  fmt.Sprintf("Function access granted to the provided %s", style.Pluralize("user", "users", len(users))),
				})))
			case "remove":
				err = clients.API().FunctionDistributionRemoveUsers(ctx, functionFlag, app.AppID, users)
				if err != nil {
					return err
				}
				_, _ = clients.IO.WriteOut().Write([]byte(style.Sectionf(style.TextSection{
					Emoji: "firecracker",
					Text:  fmt.Sprintf("Function access revoked for the provided %s", style.Pluralize("user", "users", len(users))),
				})))
			}
			return printDistribution(ctx, cmd, clients, app)
		}

		_, err = clients.API().FunctionDistributionSet(ctx, functionFlag, app.AppID, dist, "")
		if err != nil {
			return err
		}
		return printDistribution(ctx, cmd, clients, app)
	}
}

// distributePermissionFile uses data in fileName to set function permissions
func distributePermissionFile(ctx context.Context, clients *shared.ClientFactory, app types.App, fileName string) error {
	var data types.FunctionPermissions
	if file, err := afero.ReadFile(clients.Fs, fileName); err != nil {
		return err
	} else {
		extension := filepath.Ext(fileName)
		switch strings.ToLower(extension) {
		case ".json":
			if err := json.Unmarshal([]byte(file), &data); err != nil {
				return err
			}
		case ".yaml", ".yml":
			if err := yaml.Unmarshal([]byte(file), &data); err != nil {
				return err
			}
		default:
			return slackerror.New(slackerror.ErrUnknownFileType).
				WithMessage("The file extension '%s' is unknown", extension).
				WithRemediation("Expected file types include '.json' or '.yaml'")
		}
	}
	for function, permissions := range data.FunctionMap {
		if !permissions.Type.IsValid() {
			return slackerror.New(slackerror.ErrInvalidPermissionType).
				WithMessage("An unexpected permission type was provided").
				WithDetails(slackerror.ErrorDetails{
					slackerror.ErrorDetail{
						Message: fmt.Sprintf(
							"The type '%s' is not valid for function '%s'",
							permissions.Type,
							function,
						),
						Remediation: fmt.Sprintf(
							"Replace it with '%s', '%s', or '%s'",
							types.PermissionEveryone,
							types.PermissionAppCollaborators,
							types.PermissionNamedEntities,
						),
					},
				})
		}
		switch permissions.Type {
		case types.PermissionNamedEntities:
			if len(permissions.UserIDs) == 0 {
				clients.IO.PrintWarning(ctx, fmt.Sprintf(
					"No users will have access to '%s'",
					function,
				))
			}
			err := updateNamedEntitiesDistribution(ctx, clients, app, function, permissions.UserIDs)
			if err != nil {
				return err
			}
		default:
			if len(permissions.UserIDs) != 0 {
				clients.IO.PrintWarning(ctx, fmt.Sprintf(
					"The supplied user IDs to '%s' are overridden by the '%s' permission",
					function,
					permissions.Type,
				))
			}
			_, err := clients.API().FunctionDistributionSet(ctx, function, app.AppID, permissions.Type, "")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// updateNamedEntitiesDistribution removes unspecified entities and adds any
// specified ones to a function
func updateNamedEntitiesDistribution(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	function string,
	entities []string,
) error {
	updatedUsers := strings.Join(entities, ",")
	_, err := clients.API().FunctionDistributionSet(ctx, function, app.AppID, types.PermissionNamedEntities, updatedUsers)
	if err != nil {
		return err
	}
	_, currentUsers, err := clients.API().FunctionDistributionList(ctx, function, app.AppID)
	if err != nil {
		return err
	}
	removalSlice := []string{}
	for _, currentUser := range currentUsers {
		contains := false
		for _, updatedUser := range entities {
			if updatedUser == currentUser.ID {
				contains = true
				continue
			}
		}
		if !contains {
			removalSlice = append(removalSlice, currentUser.ID)
		}
	}
	removals := strings.Join(removalSlice, ",")
	err = clients.API().FunctionDistributionRemoveUsers(ctx, function, app.AppID, removals)
	if err != nil {
		return err
	}
	return nil
}

// printDistribution formats and displays access information
func printDistribution(ctx context.Context, cmd *cobra.Command, clients *shared.ClientFactory, app types.App) error {
	dist, userAccessList, err := clients.API().FunctionDistributionList(ctx, functionFlag, app.AppID)
	if err != nil {
		return err
	}
	var entities string
	if dist == types.PermissionAppCollaborators {
		entities = "app collaborators"
	} else {
		entities = "the following users"
	}
	var emoji string
	var secondary []string
	switch {
	case dist == types.PermissionEveryone:
		emoji = "busts_in_silhouette"
		secondary = append(secondary, types.GetAccessTypeDescriptionForEveryone(app))
	case len(userAccessList) <= 0:
		emoji = "ghost"
		secondary = append(secondary, style.Secondary("list is empty"))
	default:
		emoji = "bust_in_silhouette"
		for _, entity := range userAccessList {
			userInfo := []string{entity.ID}
			if entity.Email != "" {
				userInfo = append(userInfo, entity.Email)
			}
			secondary = append(secondary, fmt.Sprintf("%s (%s)", entity.UserName, strings.Join(userInfo, ", ")))
		}
	}
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     emoji,
		Text:      fmt.Sprintf("Function '%s' can be added to workflows by %s:", functionFlag, entities),
		Secondary: secondary,
	}))
	return nil
}

// printEntityAccess formats and displays the users with access to an app's functions
func printEntityAccess(ctx context.Context, cmd *cobra.Command, clients *shared.ClientFactory, app types.App) error {
	distType, users, err := clients.API().FunctionDistributionList(ctx, functionFlag, app.AppID)
	if err != nil {
		return err
	}
	if distType != types.PermissionNamedEntities {
		return nil
	}
	var emoji string
	var text string
	var secondary []string
	if len(users) > 0 {
		emoji = "bust_in_silhouette"
		text = "The following users currently have access"
		for _, u := range users {
			userInfo := []string{u.ID}
			if u.Email != "" {
				userInfo = append(userInfo, u.Email)
			}
			secondary = append(secondary, fmt.Sprintf("%s (%s)", u.UserName, strings.Join(userInfo, ", ")))
		}
	} else {
		emoji = "ghost"
		text = "No one has access yet"
	}
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     emoji,
		Text:      text,
		Secondary: secondary,
	}))
	return nil
}

// handleDistributionType checks if the function's distribution type is named_entities and if not, updates it
func handleDistributionType(ctx context.Context, clients *shared.ClientFactory, app types.App) error {
	distType, _, err := clients.API().FunctionDistributionList(ctx, functionFlag, app.AppID)
	if err != nil {
		return err
	}

	if distType == types.PermissionNamedEntities {
		return nil
	}

	_, err = clients.API().FunctionDistributionSet(ctx, functionFlag, app.AppID, types.PermissionNamedEntities, "")
	if err != nil {
		return err
	}

	return nil
}

// AddCollaboratorsToNamedEntities is a convenience method to let collaborators maintain their access when distribution type is changed from app_collaborators to named_entities
func AddCollaboratorsToNamedEntities(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	token string,
) error {
	collaborators, err := clients.API().ListCollaborators(ctx, token, app.AppID)
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

	err = handleDistributionType(ctx, clients, app)
	if err != nil {
		return err
	}

	return clients.API().FunctionDistributionAddUsers(ctx, functionFlag, app.AppID, userIDs)
}
