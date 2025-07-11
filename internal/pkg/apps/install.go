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

package apps

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/manifest"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

// Constants for onlyCreateUpdateAppManifest parameter
const (
	CreateAppManifestOnly       = true
	CreateAppManifestAndInstall = false
)

const additionalManifestInfoNotice = "App manifest contains some components that may require additional information"

// Install installs the app to a team
func Install(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, auth types.SlackAuth, onlyCreateUpdateAppManifest bool, app types.App, orgGrantWorkspaceID string) (types.App, types.InstallState, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pkg.apps.install")
	defer span.Finish()

	manifestUpdates, err := shouldUpdateManifest(ctx, clients, app, auth)
	if err != nil {
		return types.App{}, "", err
	}
	manifestCreates, err := shouldCreateManifest(ctx, clients, app)
	if err != nil {
		return types.App{}, "", err
	}

	if !clients.Config.WithExperimentOn(experiment.BoltInstall) {
		if !manifestUpdates && !manifestCreates {
			return app, "", nil
		}
	}

	// Get the token for the authenticated workspace
	apiInterface := clients.API()
	token := auth.Token
	authSession, err := apiInterface.ValidateSession(ctx, token)
	if err != nil {
		return types.App{}, "", slackerror.Wrap(err, slackerror.ErrInvalidAuth)
	}

	// Set the user_id, team id, team_domain of team that app belongs to on context
	// TODO: we should probably pick one place to store team/user/enterprise ID
	ctx = config.SetContextTeamID(ctx, *authSession.TeamID)
	clients.EventTracker.SetAuthTeamID(*authSession.TeamID)
	ctx = config.SetContextTeamDomain(ctx, auth.TeamDomain)
	if authSession.UserID != nil {
		ctx = config.SetContextUserID(ctx, *authSession.UserID)
		clients.EventTracker.SetAuthUserID(*authSession.UserID)
	}
	if authSession.EnterpriseID != nil {
		config.SetContextEnterpriseID(ctx, *authSession.EnterpriseID)
		clients.EventTracker.SetAuthEnterpriseID(*authSession.EnterpriseID)
		app.EnterpriseID = *authSession.EnterpriseID
	}

	// When the BoltInstall experiment is enabled, we need to get the manifest from the local file
	// if the manifest source is local or if we are creating a new app. After an app is created,
	// app settings becomes the source of truth for remote manifests, so updates and installs always
	// get the latest manifest from app settings.
	// When the BoltInstall experiment is disabled, we get the manifest from the local file because
	// this is how the original implementation worked.
	var slackManifest types.SlackYaml
	if clients.Config.WithExperimentOn(experiment.BoltInstall) {
		manifestSource, err := clients.Config.ProjectConfig.GetManifestSource(ctx)
		if err != nil {
			return app, "", err
		}
		if manifestSource.Equals(config.ManifestSourceLocal) || manifestCreates {
			slackManifest, err = clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
			if err != nil {
				return app, "", err
			}
		} else {
			slackManifest, err = clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
			if err != nil {
				return app, "", err
			}
		}
	} else {
		slackManifest, err = clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
		if err != nil {
			return app, "", err
		}
	}

	log.Data["appName"] = slackManifest.DisplayInformation.Name
	log.Data["isUpdate"] = app.AppID != ""
	log.Data["teamName"] = *authSession.TeamName
	log.Log("INFO", "app_install_manifest")

	manifest := slackManifest.AppManifest
	if slackManifest.IsFunctionRuntimeSlackHosted() {
		configureHostedManifest(ctx, clients, &manifest)
	}

	err = validateManifestForInstall(ctx, clients, token, app, manifest)
	if err != nil {
		return app, "", err
	}

	start := time.Now()
	switch {
	case manifestUpdates:
		log.Info("app_install_manifest_update")
		clients.IO.PrintDebug(ctx, "updating app %s", app.AppID)
		_, err := apiInterface.UpdateApp(ctx, token, app.AppID, manifest, clients.Config.ForceFlag, true)
		if err != nil {
			return app, "", err
		}
	case manifestCreates:
		log.Info("app_install_manifest_create")
		clients.IO.PrintDebug(ctx, "app not found so creating a new app")
		result, err := apiInterface.CreateApp(ctx, token, manifest, false)
		if err != nil {
			err = slackerror.Wrap(err, slackerror.ErrAppInstall)
			return app, "", err
		}
		clients.IO.PrintDebug(ctx, "created new app ID %s", result.AppID)

		// Set app properties
		app.AppID = result.AppID
		app.TeamID = *authSession.TeamID
		app.TeamDomain = auth.TeamDomain
		// TODO: add enterprise ID and user ID to app? See InstallLocalApp.
		// app.EnterpriseID = config.GetContextEnterpriseID(ctx)
		// app.UserID = *authSession.UserID
	}

	appManageURL := fmt.Sprintf("%s/apps", apiInterface.Host())
	log.Data["appURL"] = fmt.Sprintf("%s%s", appManageURL, app.AppID)
	log.Data["appName"] = manifest.DisplayInformation.Name

	if !clients.Config.SkipLocalFs() {
		if err := clients.AppClient().SaveDeployed(ctx, app); err != nil {
			return types.App{}, "", err
		}
	}
	caches, err := shouldCacheManifest(ctx, clients, app)
	if err != nil {
		return types.App{}, "", err
	}
	if caches {
		saved, err := clients.Config.ProjectConfig.Cache().GetManifestHash(ctx, app.AppID)
		if err != nil {
			return types.App{}, "", err
		}
		upstream, err := clients.API().ExportAppManifest(ctx, auth.Token, app.AppID)
		if err != nil {
			return types.App{}, "", err
		}
		hash, err := clients.Config.ProjectConfig.Cache().NewManifestHash(ctx, upstream.Manifest.AppManifest)
		if err != nil {
			return types.App{}, "", err
		}
		if !hash.Equals(saved) {
			err := clients.Config.ProjectConfig.Cache().SetManifestHash(ctx, app.AppID, hash)
			if err != nil {
				return types.App{}, "", err
			}
		}
	}

	// Install the app to a workspace
	if onlyCreateUpdateAppManifest {
		return app, "", nil
	}

	botScopes := []string{}
	if manifest.OAuthConfig != nil {
		botScopes = manifest.OAuthConfig.Scopes.Bot
	}

	outgoingDomains := []string{}
	if manifest.OutgoingDomains != nil {
		outgoingDomains = *manifest.OutgoingDomains
	}

	log.Info("app_install_start")
	// Note - we use DeveloperAppInstall endpoint for both local (dev) runs
	// and hosted installs https://github.com/slackapi/slack-cli/pull/456#discussion_r830272175

	result, installState, err := apiInterface.DeveloperAppInstall(ctx, clients.IO, token, app, botScopes, outgoingDomains, orgGrantWorkspaceID, clients.Config.AutoRequestAAAFlag)
	if err != nil {
		err = slackerror.Wrap(err, slackerror.ErrAppInstall)
		return app, "", err
	}

	if installState != types.InstallSuccess {
		printNonSuccessInstallState(ctx, clients, installState)
		return app, installState, nil
	}

	if manifest.FunctionRuntime() != types.SlackHosted {
		if err := setAppEnvironmentTokens(ctx, clients, result); err != nil {
			return app, installState, err
		}
	}

	// upload icon, default to icon.png
	var iconPath = slackManifest.Icon
	if iconPath == "" {
		if _, err := os.Stat("icon.png"); !os.IsNotExist(err) {
			iconPath = "icon.png"
		}
	}
	if iconPath != "" {
		log.Data["iconPath"] = iconPath
		err = updateIcon(ctx, clients, iconPath, app.AppID, token)
		if err != nil {
			clients.IO.PrintDebug(ctx, "icon error: %s", err)
			log.Data["iconError"] = err.Error()
			log.Info("app_install_icon_error")
		} else {
			log.Info("app_install_icon_success")
		}
		// TODO: Optimization.
		// Save a md5 hash of the icon in environments.yaml
		// var iconHash string
		// iconHash, err = getIconHash(iconPath)
		// if err != nil {
		// 	return env, err
		// }
		// env.IconHash = iconHash
	}
	// update config with latest yaml hash
	// env.Hash = slackYaml.Hash

	log.Data["installTime"] = fmt.Sprintf("%.1fs", time.Since(start).Seconds())
	log.Info("app_install_complete")

	return app, types.InstallSuccess, nil
}

func printNonSuccessInstallState(ctx context.Context, clients *shared.ClientFactory, installState types.InstallState) {
	var (
		primary   string
		secondary string
	)
	switch installState {
	case types.InstallRequestPending:
		primary = "Your request to install the app is pending"
		secondary = fmt.Sprintf("You will receive a Slackbot message after an admin has reviewed your request\nOnce your request is approved, complete installation by re-running %s", style.Commandf(clients.Config.Command, true))
	case types.InstallRequestCancelled:
		primary = "Your request to install the app has been cancelled"
		secondary = ""
	case types.InstallRequestNotSent:
		primary = "You've declined to send a request to an admin"
		secondary = "Please submit a request to install or update your app"
	}
	var status = fmt.Sprintf("\n%s", style.Sectionf(style.TextSection{
		Emoji: "bell",
		Text:  primary,
		Secondary: []string{
			secondary,
		},
	}))
	clients.IO.PrintInfo(ctx, false, status)
}

func validateManifestForInstall(ctx context.Context, clients *shared.ClientFactory, token string, app types.App, appManifest types.AppManifest) error {
	validationResult, err := clients.API().ValidateAppManifest(ctx, token, appManifest, app.AppID)

	if retryValidate := manifest.HandleConnectorNotInstalled(ctx, clients, token, err); retryValidate {
		validationResult, err = clients.API().ValidateAppManifest(ctx, token, appManifest, app.AppID)
	}

	if err := manifest.HandleConnectorApprovalRequired(ctx, clients, token, err); err != nil {
		return err
	}

	// The apps.manifest.validate API returns both breaking changes and warnings in the warnings key of the API response.
	// Checking here to see if there are any warnings because we still want to show warnings for new apps.
	foundWarning := false
	for _, s := range validationResult.Warnings {
		if s.Code != "breaking_change" {
			foundWarning = true
			break
		}
	}

	// creating new app so simply return any error or null
	if app.AppID == "" && !foundWarning {
		return err
	}

	// it is an update so additional compatibility/breaking-changes checks required
	if err != nil {
		isDatastoreSchemaCompatibilityError := strings.Contains(err.Error(), "schema_compatibility_error")
		if !isDatastoreSchemaCompatibilityError {
			// it is not a datastore schema compatibility error so return error
			clients.IO.PrintDebug(ctx, "failed updating app %s: %s", app.AppID, err)
			return err
		}

		if !clients.Config.ForceFlag {
			var commandSuggestion string
			if app.IsDev {
				commandSuggestion = style.Commandf("run --force", false)
			} else {
				commandSuggestion = style.Commandf("deploy --force", false)
			}
			// it is a datastore schema compatibility error, warn and exit if no force flag used
			clients.IO.PrintWarning(ctx, "Proceed with %s to update your app.", commandSuggestion)
			return err
		}
	}

	// NOTE: Validations against the dev installed (`slack run`) version of the app and the deployed (`slack deploy`) version
	// could be entirely different. Also, because the `manifest validate` command compares to the prod version it's possible to run
	// these commands in sequence and have one return warnings but not the other (or both return different warnings).
	warnings := validationResult.Warnings
	continueWithBreakingChanges := clients.Config.ForceFlag
	if len(warnings) > 0 {
		if !clients.Config.ForceFlag {
			continueWithBreakingChanges, err = continueDespiteWarning(ctx, clients, warnings)
			if err != nil {
				return err
			}
		}
		if !continueWithBreakingChanges {
			return slackerror.New("Cancelled due to user input")
		}
	}

	return nil
}

// InstallLocalApp installs a non-hosted local app to a workspace.
func InstallLocalApp(ctx context.Context, clients *shared.ClientFactory, orgGrantWorkspaceID string, log *logger.Logger, auth types.SlackAuth, app types.App) (types.App, api.DeveloperAppInstallResult, types.InstallState, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "installLocalApp")
	defer span.Finish()

	manifestUpdates, err := shouldUpdateManifest(ctx, clients, app, auth)
	if err != nil {
		return types.App{}, api.DeveloperAppInstallResult{}, "", err
	}
	manifestCreates, err := shouldCreateManifest(ctx, clients, app)
	if err != nil {
		return types.App{}, api.DeveloperAppInstallResult{}, "", err
	}

	if !clients.Config.WithExperimentOn(experiment.BoltInstall) {
		if !manifestUpdates && !manifestCreates {
			return app, api.DeveloperAppInstallResult{}, "", nil
		}
	}

	apiInterface := clients.API()
	token := auth.Token
	authSession, err := apiInterface.ValidateSession(ctx, token)
	if err != nil {
		return app, api.DeveloperAppInstallResult{}, "", slackerror.Wrap(err, slackerror.ErrInvalidAuth)
	}

	// Set the user_id, team id, team_domain of team that app belongs to on context
	// TODO: we should probably pick one place to store team/user/enterprise ID
	ctx = config.SetContextTeamID(ctx, *authSession.TeamID)
	clients.EventTracker.SetAuthTeamID(*authSession.TeamID)
	ctx = config.SetContextTeamDomain(ctx, auth.TeamDomain)
	if authSession.UserID != nil {
		ctx = config.SetContextUserID(ctx, *authSession.UserID)
		clients.EventTracker.SetAuthUserID(*authSession.UserID)
	}
	if authSession.EnterpriseID != nil {
		ctx = config.SetContextEnterpriseID(ctx, *authSession.EnterpriseID)
		clients.EventTracker.SetAuthEnterpriseID(*authSession.EnterpriseID)
		// TODO: add enterprise ID to app? See Install.
		// app.EnterpriseID = *authSession.EnterpriseID
	}

	// When the BoltInstall experiment is enabled, we need to get the manifest from the local file
	// if the manifest source is local or if we are creating a new app. After an app is created,
	// app settings becomes the source of truth for remote manifests, so updates and installs always
	// get the latest manifest from app settings.
	// When the BoltInstall experiment is disabled, we get the manifest from the local file because
	// this is how the original implementation worked.
	var slackManifest types.SlackYaml
	if clients.Config.WithExperimentOn(experiment.BoltInstall) {
		manifestSource, err := clients.Config.ProjectConfig.GetManifestSource(ctx)
		if err != nil {
			return app, api.DeveloperAppInstallResult{}, "", err
		}
		if manifestSource.Equals(config.ManifestSourceLocal) || manifestCreates {
			slackManifest, err = clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
			if err != nil {
				return app, api.DeveloperAppInstallResult{}, "", err
			}
		} else {
			slackManifest, err = clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
			if err != nil {
				return app, api.DeveloperAppInstallResult{}, "", err
			}
		}
	} else {
		slackManifest, err = clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
		if err != nil {
			return app, api.DeveloperAppInstallResult{}, "", err
		}
	}

	log.Data["appName"] = slackManifest.DisplayInformation.Name
	log.Data["isUpdate"] = app.AppID != ""
	log.Data["teamName"] = *authSession.TeamName
	log.Log("INFO", "app_install_manifest")

	manifest := slackManifest.AppManifest
	appendLocalToDisplayName(&manifest)
	if manifest.IsFunctionRuntimeSlackHosted() {
		configureLocalManifest(ctx, clients, &manifest)
	}

	err = validateManifestForInstall(ctx, clients, token, app, manifest)
	if err != nil {
		return app, api.DeveloperAppInstallResult{}, "", err
	}

	start := time.Now()
	switch {
	case manifestUpdates:
		log.Info("app_install_manifest_update")
		log.Info("on_update_app_install")
		clients.IO.PrintDebug(ctx, "updating app %s", app.AppID)
		_, err := apiInterface.UpdateApp(ctx, token, app.AppID, manifest, clients.Config.ForceFlag, true)
		if err != nil {
			clients.IO.PrintDebug(ctx, "failed updating app %s: %s", app.AppID, err)
			return app, api.DeveloperAppInstallResult{}, "", err
		}
	case manifestCreates:
		log.Info("app_install_manifest_create")
		clients.IO.PrintDebug(ctx, "app not found so creating a new app")
		result, err := apiInterface.CreateApp(ctx, token, manifest, false)
		if err != nil {
			err = slackerror.Wrap(err, slackerror.ErrAppInstall)
			return app, api.DeveloperAppInstallResult{}, "", err
		}
		clients.IO.PrintDebug(ctx, "created new app ID %s", result.AppID)

		// Set properties on app
		app.AppID = result.AppID
		// TODO: should we be using context to store this information? seems risky
		app.TeamID = config.GetContextTeamID(ctx)
		app.TeamDomain = config.GetContextTeamDomain(ctx)
		app.EnterpriseID = config.GetContextEnterpriseID(ctx)
		app.UserID = *authSession.UserID
	}

	appManageURL := fmt.Sprintf("%s/apps", apiInterface.Host())
	log.Data["appURL"] = fmt.Sprintf("%s%s", appManageURL, app.AppID)
	log.Data["appName"] = manifest.DisplayInformation.Name

	// specifically set app.IsDev to be true for dev installation
	app.IsDev = true

	// save updated or created app to apps.dev.json
	if !clients.Config.SkipLocalFs() {
		if err := clients.AppClient().SaveLocal(ctx, app); err != nil {
			return types.App{}, api.DeveloperAppInstallResult{}, "", err
		}
	}
	caches, err := shouldCacheManifest(ctx, clients, app)
	if err != nil {
		return types.App{}, api.DeveloperAppInstallResult{}, "", err
	}
	if caches {
		saved, err := clients.Config.ProjectConfig.Cache().GetManifestHash(ctx, app.AppID)
		if err != nil {
			return types.App{}, api.DeveloperAppInstallResult{}, "", err
		}
		upstream, err := clients.API().ExportAppManifest(ctx, auth.Token, app.AppID)
		if err != nil {
			return types.App{}, api.DeveloperAppInstallResult{}, "", err
		}
		hash, err := clients.Config.ProjectConfig.Cache().NewManifestHash(ctx, upstream.Manifest.AppManifest)
		if err != nil {
			return types.App{}, api.DeveloperAppInstallResult{}, "", err
		}
		if !hash.Equals(saved) {
			err := clients.Config.ProjectConfig.Cache().SetManifestHash(ctx, app.AppID, hash)
			if err != nil {
				return types.App{}, api.DeveloperAppInstallResult{}, "", err
			}
		}
	}

	// install the app
	var botScopes []string
	if manifest.OAuthConfig != nil {
		botScopes = manifest.OAuthConfig.Scopes.Bot
	}

	outgoingDomains := []string{}
	if manifest.OutgoingDomains != nil {
		outgoingDomains = *manifest.OutgoingDomains
	}

	log.Info("app_install_start")
	var installState types.InstallState
	result, installState, err := apiInterface.DeveloperAppInstall(ctx, clients.IO, token, app, botScopes, outgoingDomains, orgGrantWorkspaceID, clients.Config.AutoRequestAAAFlag)

	if err != nil {
		err = slackerror.Wrap(err, slackerror.ErrAppInstall)
		return app, api.DeveloperAppInstallResult{}, "", err
	}

	if installState != types.InstallSuccess {
		printNonSuccessInstallState(ctx, clients, installState)
		return app, api.DeveloperAppInstallResult{}, installState, nil
	}

	if err := setAppEnvironmentTokens(ctx, clients, result); err != nil {
		return app, result, installState, err
	}

	//
	// TODO: Currently, cannot update the icon if app is not hosted.
	//
	// upload icon, default to icon.png
	// var iconPath = slackYaml.Icon
	// if iconPath == "" {
	// 	if _, err := os.Stat("icon.png"); !os.IsNotExist(err) {
	// 		iconPath = "icon.png"
	// 	}
	// }
	// if iconPath != "" {
	// 	clients.IO.PrintDebug(ctx, "uploading icon")
	// 	err = updateIcon(ctx, clients, iconPath, env.AppID, token)
	// 	if err != nil {
	// 		clients.IO.PrintError(ctx, "An error occurred updating the Icon", err)
	// 	}
	// 	// Save a md5 hash of the icon in environments.yaml
	// 	var iconHash string
	// 	iconHash, err = getIconHash(iconPath)
	// 	if err != nil {
	// 		return env, api.DeveloperAppInstallResult{}, err
	// 	}
	// 	env.IconHash = iconHash
	// }

	// update config with latest yaml hash
	// env.Hash = slackYaml.Hash

	log.Data["installTime"] = fmt.Sprintf("%.1fs", time.Since(start).Seconds())
	log.Info("app_install_complete")

	return app, result, types.InstallSuccess, nil
}

// getIconHash returns the MD5 hash of the icon file
// func getIconHash(iconPath string) (string, error) {
// 	if iconPath == "" {
// 		return "", slackerror.New("missing required args")
// 	}

// 	var icon *os.File
// 	var err error
// 	icon, err = os.Open(iconPath)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer icon.Close()

// 	var hash = md5.New()
// 	if _, err := io.Copy(hash, icon); err != nil {
// 		return "", err
// 	}
// 	return hex.EncodeToString(hash.Sum(nil)), nil
// }

// configureHostedManifest sets the expected manifest values for hosted runtimes
//
// Run On Slack apps have certain runtime requirements so certain values are set
// here before installing the application. This includes interactivity and event
// subscriptions as specified through Run On Slack hosting requirements.
//
// The CLI determines the API host from selected credentials during installation.
//
// Apps without a specified or a "remote" function runtime should ignore this.
func configureHostedManifest(
	ctx context.Context,
	clients *shared.ClientFactory,
	manifest *types.AppManifest,
) {
	if manifest == nil {
		return
	}
	clients.IO.PrintDebug(ctx, "updating app manifest with required properties for a run-on-slack function runtime")
	if manifest.Settings == nil {
		manifest.Settings = &types.AppSettings{}
	}
	manifest.Settings.FunctionRuntime = types.SlackHosted
	if manifest.Settings.Interactivity == nil {
		manifest.Settings.Interactivity = &types.ManifestInteractivity{}
	}
	host := clients.API().Host()
	manifest.Settings.Interactivity.IsEnabled = true
	manifest.Settings.Interactivity.MessageMenuOptionsURL = host
	manifest.Settings.Interactivity.RequestURL = host
	if manifest.Settings.EventSubscriptions == nil {
		manifest.Settings.EventSubscriptions = &types.ManifestEventSubscriptions{}
	}
	manifest.Settings.EventSubscriptions.RequestURL = host
}

// configureLocalManifest sets the default manifest values for local runtimes
//
// The "local" runtime applies to just Run On Slack apps and this configuration
// is a workaround for undefined values from the "get-manifest" hook as provided
// through the Deno SDK default Manifest export.
//
// This case is known since that "get-manifest" hook in the Deno SDK also sets
// the function runtime to "slack". If this changes, the next semver:major can
// remove this case and require a minimum Deno SDK version. For now it allows
// quick development with Socket Mode.
//
// Apps without a specified or a "remote" function runtime should ignore this.
func configureLocalManifest(
	ctx context.Context,
	clients *shared.ClientFactory,
	manifest *types.AppManifest,
) {
	if manifest.Settings == nil {
		manifest.Settings = &types.AppSettings{}
	}
	clients.IO.PrintDebug(ctx, "updating app manifest with default properties for a run-on-slack function runtime")
	manifest.Settings.FunctionRuntime = types.LocallyRun
	t := true
	manifest.Settings.SocketModeEnabled = &t
	if manifest.Settings.Interactivity == nil {
		manifest.Settings.Interactivity = &types.ManifestInteractivity{}
	}
	manifest.Settings.Interactivity.IsEnabled = true
	manifest.Settings.Interactivity.RequestURL = ""
	manifest.Settings.Interactivity.MessageMenuOptionsURL = ""
	if manifest.Settings.EventSubscriptions == nil {
		manifest.Settings.EventSubscriptions = &types.ManifestEventSubscriptions{}
	}
	manifest.Settings.EventSubscriptions.RequestURL = ""
}

// appendLocalToDisplayName appends a "(local)" tag to the application names
func appendLocalToDisplayName(manifest *types.AppManifest) {
	manifest.DisplayInformation.Name = style.LocalRunDisplayName(manifest.DisplayInformation.Name)
	if manifest.Features != nil {
		manifest.Features.BotUser.DisplayName = style.LocalRunDisplayName(manifest.Features.BotUser.DisplayName)
	}
}

// updateIcon will upload the new icon to the Slack API
func updateIcon(ctx context.Context, clients *shared.ClientFactory, iconPath, appID string, token string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "updateIcon")
	defer span.Finish()

	clients.IO.PrintDebug(ctx, "uploading icon")

	// var iconResp apiclient.IconResult
	var err error
	_, err = clients.API().Icon(ctx, clients.Fs, token, appID, iconPath)
	if err != nil {
		// TODO: separate the icon upload into a different function because if an error is returned
		// the new app_id might be ignored and next time we'll create another app.
		return fmt.Errorf("%s %s", err, iconPath)
	}

	// Save a md5 hash of the icon in environments.yaml
	// env.IconHash = iconResp.MD5Hash
	return nil
}

// shouldCreateManifest decides if an app manifest needs to be created for an
// app to exist
func shouldCreateManifest(ctx context.Context, clients *shared.ClientFactory, app types.App) (bool, error) {
	if !clients.Config.WithExperimentOn(experiment.BoltFrameworks) {
		return app.AppID == "", nil
	}

	// When the BoltInstall experiment is enabled, apps can always be created with any manifest source.
	if clients.Config.WithExperimentOn(experiment.BoltInstall) {
		return app.AppID == "", nil
	}

	manifestSource, err := clients.Config.ProjectConfig.GetManifestSource(ctx)
	if err != nil {
		return false, err
	}
	return app.AppID == "" && manifestSource == config.ManifestSourceLocal, nil
}

// shouldCacheManifest decides if an app manifest hash should be saved to cache
func shouldCacheManifest(ctx context.Context, clients *shared.ClientFactory, app types.App) (bool, error) {
	if !clients.Config.WithExperimentOn(experiment.BoltFrameworks) {
		return false, nil
	}
	manifestSource, err := clients.Config.ProjectConfig.GetManifestSource(ctx)
	if err != nil {
		return false, err
	}
	if manifestSource.Equals(config.ManifestSourceRemote) {
		return false, nil
	}
	manifest, err := clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
	if err != nil {
		return false, err
	}
	if manifest.IsFunctionRuntimeSlackHosted() {
		return false, nil
	}
	saved, err := clients.Config.ProjectConfig.Cache().GetManifestHash(ctx, app.AppID)
	if err != nil {
		return false, err
	}
	if saved != "" {
		return true, nil
	}
	if clients.Config.SkipLocalFs() {
		return false, nil
	}
	return true, nil
}

// shouldUpdateManifest decides if an existing app manifest should be updated
func shouldUpdateManifest(ctx context.Context, clients *shared.ClientFactory, app types.App, auth types.SlackAuth) (bool, error) {
	if app.AppID == "" {
		return false, nil
	}
	if !clients.Config.WithExperimentOn(experiment.BoltFrameworks) {
		return true, nil
	}
	manifestSource, err := clients.Config.ProjectConfig.GetManifestSource(ctx)
	if err != nil {
		return false, err
	}
	if manifestSource.Equals(config.ManifestSourceRemote) {
		return false, nil
	}
	if clients.Config.ForceFlag {
		return true, nil
	}
	manifest, err := clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
	if err != nil {
		return false, err
	}
	if manifest.IsFunctionRuntimeSlackHosted() {
		return true, nil
	}
	saved, err := clients.Config.ProjectConfig.Cache().GetManifestHash(ctx, app.AppID)
	if err != nil {
		return false, err
	}
	upstream, err := clients.API().ExportAppManifest(ctx, auth.Token, app.AppID)
	if err != nil {
		return false, err
	}
	hash, err := clients.Config.ProjectConfig.Cache().NewManifestHash(ctx, upstream.Manifest.AppManifest)
	if err != nil {
		return false, err
	}
	notice := ""
	switch {
	case saved.Equals(hash):
		return true, nil
	case saved.Equals(""):
		notice = "Manifest values for this app are overwritten on reinstall"
	default:
		notice = "The manifest on app settings has been changed since last update!"
	}
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "books",
		Text:      "App Manifest",
		Secondary: []string{notice},
	}))
	if !clients.IO.IsTTY() {
		return false, errorAppManifestUpdate(app, true)
	}
	continues, err := clients.IO.ConfirmPrompt(
		ctx,
		fmt.Sprintf(
			"Update app settings with changes to the %s manifest?",
			config.ManifestSourceLocal.String(),
		),
		false,
	)
	if err != nil {
		return false, err
	}
	if !continues {
		return false, errorAppManifestUpdate(app, false)
	}
	return true, nil
}

// errorAppManifestUpdate formats an error message with app specific remediation
func errorAppManifestUpdate(app types.App, forceOption bool) *slackerror.Error {
	url := "https://api.slack.com/apps"
	switch {
	case app.AppID != "" && app.EnterpriseID != "":
		url = fmt.Sprintf("https://app.slack.com/app-settings/%s/%s/app-manifest", app.EnterpriseID, app.AppID)
	case app.AppID != "" && app.TeamID != "":
		url = fmt.Sprintf("https://app.slack.com/app-settings/%s/%s/app-manifest", app.TeamID, app.AppID)
	case app.AppID != "":
		url = fmt.Sprintf("https://api.slack.com/apps/%s", app.AppID)
	}
	command := style.Commandf(fmt.Sprintf("manifest --source %s", config.ManifestSourceLocal.String()), false)
	remediation := []string{
		fmt.Sprintf("Check %s values with %s", config.ManifestSourceLocal.String(), command),
		fmt.Sprintf("Compare app settings: %s", style.LinkText(url)),
	}
	if forceOption {
		option := fmt.Sprintf("Write %s manifest values to app settings using `%s`",
			config.ManifestSourceLocal.String(),
			style.CommandText("--force"),
		)
		remediation = append(remediation, option)
	}
	return slackerror.New(slackerror.ErrAppManifestUpdate).WithRemediation("%s", strings.Join(remediation, "\n"))
}

// Displays warning message and details and then gives user a prompt on if they want to continue or not. Returns false
// if user does NOT wish to proceed, true otherwise.
func continueDespiteWarning(ctx context.Context, clients *shared.ClientFactory, warn slackerror.Warnings) (bool, error) {
	// Some warnings can just be warnings and not breaking change alerts
	// In that case, we don't want to prompt the user
	var foundBreakingChange = false
	for _, s := range warn {
		if s.Code == "breaking_change" {
			foundBreakingChange = true
			break
		}
	}

	if foundBreakingChange {
		clients.IO.PrintWarning(ctx, warn.Warning(clients.Config.DebugEnabled, "App manifest contains possible breaking changes"))
		saveManifestDespiteWarning, err := clients.IO.ConfirmPrompt(ctx, "Confirm changes?", false)
		if err != nil {
			return false, err
		}

		if saveManifestDespiteWarning {
			clients.IO.PrintInfo(ctx, false,
				"\n%s: %s",
				style.Bold("Changes confirmed"),
				style.Styler().Green("Continuing with install."),
			)
			return true, nil
		}

		clients.IO.PrintInfo(ctx, false, "\n%s", style.Styler().Red("App install canceled."))

		return false, nil
	}

	clients.IO.PrintWarning(ctx, warn.Warning(clients.Config.DebugEnabled, additionalManifestInfoNotice))

	return true, nil
}

// setAppEnvironmentTokens adds the app and bot token to the process environment
func setAppEnvironmentTokens(ctx context.Context, clients *shared.ClientFactory, result api.DeveloperAppInstallResult) error {
	if token, ok := clients.Os.LookupEnv("SLACK_APP_TOKEN"); !ok {
		if err := clients.Os.Setenv("SLACK_APP_TOKEN", result.APIAccessTokens.AppLevel); err != nil {
			return err
		}
	} else if token != result.APIAccessTokens.AppLevel {
		clients.IO.PrintWarning(ctx, style.Sectionf(style.TextSection{
			Text: fmt.Sprintf("The app token differs from the set %s environment variable", style.Highlight("SLACK_APP_TOKEN")),
			Secondary: []string{
				"The environment variable will continue to be used",
				"Proceed with caution as this might be associated to an unexpected app ID",
			},
		}))
	}
	if token, ok := clients.Os.LookupEnv("SLACK_BOT_TOKEN"); !ok {
		if err := clients.Os.Setenv("SLACK_BOT_TOKEN", result.APIAccessTokens.Bot); err != nil {
			return err
		}
	} else if token != result.APIAccessTokens.Bot {
		clients.IO.PrintWarning(ctx, style.Sectionf(style.TextSection{
			Text: fmt.Sprintf("The bot token differs from the set %s environment variable", style.Highlight("SLACK_BOT_TOKEN")),
			Secondary: []string{
				"The environment variable will continue to be used",
				"Proceed with caution as this might be associated to an unexpected bot ID",
			},
		}))
	}
	return nil
}
