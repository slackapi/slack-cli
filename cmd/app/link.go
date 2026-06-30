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

package app

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/pkg/apps"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// LinkAppConfirmPromptText is displayed when prompting to add an existing app
const LinkAppConfirmPromptText = "Do you want to add an existing app?"

// appLinkFlagSet contains flag values to reference
type appLinkFlagSet struct {
	environmentFlag string
}

// appLinkFlag contains default flag values
var appLinkFlag = appLinkFlagSet{
	environmentFlag: "",
}

// NewLinkCommand returns a new Cobra command for link
func NewLinkCommand(clients *shared.ClientFactory) *cobra.Command {
	app := &types.App{}
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Add an existing app to the project",
		Long: strings.Join([]string{
			"Saves an existing app to a project to be available to other commands.",
			"",
			"The provided App ID and Team ID are stored in the " + style.Underline("apps.json") + " or " + style.Underline("apps.dev.json"),
			"files in the .slack directory of a project.",
			"",
			"The environment option decides how an app is handled and where information",
			"should be stored. Production apps should be 'deployed' while apps used for",
			"testing and development should be considered 'local'.",
			"",
			"Only one app can exist for each combination of Team ID and environment.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Add an existing app to a project",
				Command: "app link",
			},
			{
				Meaning: "Add a specific app without using prompts",
				Command: "app link --team T0123456789 --app A0123456789 --environment deployed",
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clients.IO.PrintTrace(ctx, slacktrace.AppLinkStart)
			return LinkCommandRunE(ctx, clients, app)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clients.IO.PrintTrace(ctx, slacktrace.AppLinkSuccess)
			return nil
		},
	}
	cmd.Flags().StringVarP(&appLinkFlag.environmentFlag, "environment", "E", "", "environment to save existing app (local, deployed)")
	return cmd
}

// LinkCommandRunE saves details about the provided application
func LinkCommandRunE(ctx context.Context, clients *shared.ClientFactory, app *types.App) (err error) {
	// Add empty line between executed command and first output
	clients.IO.PrintInfo(ctx, false, "")

	err = LinkExistingApp(ctx, clients, app, false)
	if err != nil {
		return err
	}

	return nil
}

// LinkAppHeaderSection displays a section explaining how to find existing apps.
// External callers can use extraSecondaryText to show additional information.
// When shouldConfirm is true, additional information is included in the header
// explaining how to link apps, in case the user declines.
func LinkAppHeaderSection(ctx context.Context, clients *shared.ClientFactory, shouldConfirm bool) {
	var secondaryText = []string{
		"Add an existing app from app settings",
		"Find your existing apps at: " + style.Underline("https://api.slack.com/apps"),
	}

	if shouldConfirm {
		secondaryText = append(secondaryText, "Manually add apps later with "+style.Commandf("app link", true))
	}

	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
		Emoji:     "house",
		Text:      "App Link",
		Secondary: secondaryText,
	}))
}

// LinkExistingApp prompts for an existing App ID and saves the details to the project.
// When shouldConfirm is true, a confirmation prompt will ask the user if they want to
// link an existing app and additional information is included in the header.
// The shouldConfirm option is encouraged for third-party callers.
func LinkExistingApp(ctx context.Context, clients *shared.ClientFactory, app *types.App, shouldConfirm bool) (err error) {
	// Header section
	LinkAppHeaderSection(ctx, clients, shouldConfirm)

	// Confirm to add an existing app; useful for third-party callers
	if shouldConfirm {
		proceed, err := clients.IO.ConfirmPrompt(ctx, LinkAppConfirmPromptText, true)
		if err != nil {
			clients.IO.PrintDebug(ctx, "Error prompting to add an existing app: %s", err)
			return err
		}

		// Add newline to match the trailing newline inserted from the footer section
		clients.IO.PrintInfo(ctx, false, "")

		if !proceed {
			return nil
		}
	}

	// App Manifest section
	manifestSource, err := clients.Config.ProjectConfig.GetManifestSource(ctx)
	if err != nil {
		return err
	}

	configPath := filepath.Join(config.ProjectConfigDirName, config.ProjectConfigJSONFilename)
	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
		Emoji: "books",
		Text:  "App Manifest",
		Secondary: []string{
			"Manifest source is gathered from " + style.Highlight(manifestSource.Human()),
			"Manifest source is configured in " + style.Highlight(configPath),
		},
	}))

	// Prompt to get app details
	var auth *types.SlackAuth
	*app, auth, err = promptExistingApp(ctx, clients)
	if err != nil {
		return err
	}

	appIDs := []string{app.AppID}
	_, err = clients.API().GetAppStatus(ctx, auth.Token, appIDs, app.TeamID)
	if err != nil {
		return err
	}

	// Save the app to the project
	err = saveAppToJSON(ctx, clients, *app)
	if err != nil {
		clients.IO.PrintDebug(ctx, "Error saving app to file when linking existing app: %s", err)
		return err
	}

	// Footer section
	LinkAppFooterSection(ctx, clients, app)

	return nil
}

// LinkAppFooterSection displays the details of app that was added to the project.
func LinkAppFooterSection(ctx context.Context, clients *shared.ClientFactory, app *types.App) {
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "house",
		Text:      "App",
		Secondary: formatListSuccess([]types.App{*app}),
	}))
	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
		Emoji: "house_with_garden",
		Text:  "App Link",
		Secondary: []string{
			"Added existing app to project",
		},
	}))
}

// promptExistingApp gathers details to represent app information
func promptExistingApp(ctx context.Context, clients *shared.ClientFactory) (types.App, *types.SlackAuth, error) {
	slackAuth, err := prompts.PromptTeamSlackAuth(ctx, clients, "Select the existing app team", nil)
	if err != nil {
		return types.App{}, &types.SlackAuth{}, err
	}
	appID, err := promptAppID(ctx, clients)
	if err != nil {
		return types.App{}, &types.SlackAuth{}, err
	}
	isProduction, err := promptIsProduction(ctx, clients)
	if err != nil {
		return types.App{}, &types.SlackAuth{}, err
	}
	app := types.App{
		AppID:        appID,
		EnterpriseID: slackAuth.EnterpriseID,
		TeamDomain:   slackAuth.TeamDomain,
		TeamID:       slackAuth.TeamID,
	}
	if !isProduction {
		app.IsDev = true
		app.UserID = slackAuth.UserID
	}
	apps, err := apps.FetchAppInstallStates(ctx, clients, []types.App{app})
	if err != nil {
		return app, slackAuth, nil
	}
	return apps[0], slackAuth, nil
}

// promptAppID retrieves an app ID from user input
func promptAppID(ctx context.Context, clients *shared.ClientFactory) (string, error) {
	if clients.Config.Flags.Lookup("app").Changed {
		return clients.Config.Flags.Lookup("app").Value.String(), nil
	}
	value, err := clients.IO.InputPrompt(
		ctx,
		"Enter the existing app ID",
		iostreams.InputPromptConfig{
			Required: true,
		},
	)
	if err != nil {
		return "", err
	}
	return value, nil
}

// funcPromptIsProduction decides if the app should be considered production
func promptIsProduction(ctx context.Context, clients *shared.ClientFactory) (bool, error) {
	selection, err := clients.IO.SelectPrompt(
		ctx,
		"Choose the app environment",
		[]string{"Local", "Deployed"},
		iostreams.SelectPromptConfig{
			Flag:     clients.Config.Flags.Lookup("environment"),
			Required: true,
		},
	)
	if err != nil {
		return false, err
	}
	if strings.ToLower(selection.Option) == "deployed" {
		return true, nil
	} else if strings.ToLower(selection.Option) == "local" {
		return false, nil
	}
	return false, slackerror.New(slackerror.ErrMismatchedFlags).
		WithRemediation("The environment flag must be either 'local' or 'deployed'")
}

// saveAppToJSON writes the linked app to file for later use while not writing
// app IDs that exist
func saveAppToJSON(ctx context.Context, clients *shared.ClientFactory, app types.App) error {
	deploy, err := clients.AppClient().GetDeployed(ctx, app.TeamID)
	if err != nil {
		return err
	}
	local, err := clients.AppClient().GetLocal(ctx, app.TeamID)
	if err != nil {
		return err
	}
	switch app.IsDev {
	case true:
		if clients.Config.ForceFlag || (local.IsNew() && deploy.AppID != app.AppID) {
			return clients.AppClient().SaveLocal(ctx, app)
		}
	case false:
		if clients.Config.ForceFlag || (deploy.IsNew() && local.AppID != app.AppID) {
			return clients.AppClient().SaveDeployed(ctx, app)
		}
	}
	return slackerror.New(slackerror.ErrAppFound).
		WithMessage("A saved app was found and cannot be overwritten").
		WithRemediation("Remove the app from this project or try again with %s", style.Bold("--force"))
}
