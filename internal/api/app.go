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

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

const (
	appManifestCreateMethod        = "apps.manifest.create"
	appManifestExportMethod        = "apps.manifest.export"
	appManifestUpdateMethod        = "apps.manifest.update"
	appManifestValidateMethod      = "apps.manifest.validate"
	appGeneratePresignedPostMethod = "apps.hosted.generatePresignedPost"
	appUploadMethod                = "apps.hosted.upload"
	appDeleteMethod                = "apps.delete"
	appCertifiedInstallMethod      = "apps.certified.install"
	appConnectionsOpenMethod       = "apps.connections.open"
	appDeveloperInstallMethod      = "apps.developerInstall"
	appDeveloperUninstallMethod    = "apps.developerUninstall"
	appStatusMethod                = "apps.status"
	appApprovalRequestCreateMethod = "apps.approvals.requests.create"
	appApprovalRequestCancelMethod = "apps.approvals.requests.cancel"
)

// AppsClient is the interface for app-related API calls
type AppsClient interface {
	CertifiedAppInstall(ctx context.Context, token string, certifiedAppId string) (CertifiedInstallResult, error)
	ConnectionsOpen(ctx context.Context, token string) (AppsConnectionsOpenResult, error)
	CreateApp(ctx context.Context, token string, manifest types.AppManifest, enableDistribution bool) (CreateAppResult, error)
	DeleteApp(ctx context.Context, token string, appID string) error
	DeveloperAppInstall(ctx context.Context, IO iostreams.IOStreamer, token string, app types.App, botScopes []string, outgoingDomains []string, orgGrantWorkspaceID string, autoRequestAAA bool) (DeveloperAppInstallResult, types.InstallState, error)
	ExportAppManifest(ctx context.Context, token string, appId string) (ExportAppResult, error)
	GetAppStatus(ctx context.Context, token string, appIDs []string, teamID string) (GetAppStatusResult, error)
	GetPresignedS3PostParams(ctx context.Context, token string, appID string) (GenerateS3PresignedPostResult, error)
	Host() string
	Icon(ctx context.Context, fs afero.Fs, token, appID, iconFilePath string) (IconResult, error)
	RequestAppApproval(ctx context.Context, token string, appID string, teamID string, reason string, scopes string, outgoingDomains []string) (AppsApprovalsRequestsCreateResult, error)
	SetHost(host string)
	UninstallApp(ctx context.Context, token string, appID, teamID string) error
	UpdateApp(ctx context.Context, token string, appID string, manifest types.AppManifest, forceUpdate bool, continueWithBreakingChanges bool) (UpdateAppResult, error)
	UploadApp(ctx context.Context, token, runtime, appID string, fileName string) error
	UploadPackageToS3(ctx context.Context, fs afero.Fs, appID string, uploadParams GenerateS3PresignedPostResult, archiveFilePath string) (string, error)
	ValidateAppManifest(ctx context.Context, token string, manifest types.AppManifest, appId string) (ValidateAppManifestResult, error)
}

// This API returns null
type CertifiedInstallResult struct{}

type certifiedInstallResponse struct {
	extendedBaseResponse
	CertifiedInstallResult
}

// CertifiedAppInstall requests the installation of a certified app in order for its connectors to be usable
func (c *Client) CertifiedAppInstall(ctx context.Context, token string, certifiedAppId string) (CertifiedInstallResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.CertifiedAppInstall")
	defer span.Finish()

	args := struct {
		AppId string `json:"app_id,omitempty"` // the app id of the certified app
	}{
		certifiedAppId,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return CertifiedInstallResult{}, errInvalidArguments.WithRootCause(err)
	}
	b, err := c.postJSON(ctx, appCertifiedInstallMethod, token, "", body)
	if err != nil {
		return CertifiedInstallResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := certifiedInstallResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return CertifiedInstallResult{}, errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appCertifiedInstallMethod)
	}

	if !resp.Ok {
		return CertifiedInstallResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appCertifiedInstallMethod)
	}

	return resp.CertifiedInstallResult, nil

}

// App Credentials to be saved
type Credentials struct {
	ClientID          string `json:"client_id,omitempty"`
	ClientSecret      string `json:"client_secret,omitempty"`
	VerificationToken string `json:"verification_token,omitempty"`
	SigningSecret     string `json:"signing_secret,omitempty"`
}

// S3 presigned post response to be saved
type PresignedPostFields struct {
	AmzCredentials    string `json:"X-Amz-Credential"`
	AmzAlgorithm      string `json:"X-Amz-Algorithm"`
	AmzFileKey        string `json:"key"`
	AmzFileCreateDate string `json:"X-Amz-Date"`
	AmzPolicy         string `json:"Policy"`
	AmzSignature      string `json:"X-Amz-Signature"`
	AmzToken          string `json:"X-Amz-Security-Token"`
}

// CreateAppResult details to be saved
type CreateAppResult struct {
	AppID             string      `json:"app_id,omitempty"`
	Credentials       Credentials `json:"credentials,omitempty"`
	OAuthAuthorizeUrl string      `json:"oauth_authorize_url,omitempty"`
}

type createAppResponse struct {
	extendedBaseResponse
	CreateAppResult
}

var appApprovalErrorMessages = []string{
	slackerror.ErrAppApprovalRequestEligible,
	slackerror.ErrAppApprovalRequestPending,
	slackerror.ErrAppApprovalRequestDenied,
}

// CreateApp creates a new Slack app
func (c *Client) CreateApp(ctx context.Context, token string, manifest types.AppManifest, enableDistribution bool) (CreateAppResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.CreateApp")
	defer span.Finish()

	args := struct {
		Manifest           types.AppManifest `json:"manifest,omitempty"`
		EnableDistribution bool              `json:"enable_distribution,omitempty"`
	}{
		manifest,
		enableDistribution,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return CreateAppResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appManifestCreateMethod, token, "", body)
	if err != nil {
		return CreateAppResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := createAppResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return CreateAppResult{}, errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appManifestCreateMethod)
	}

	if !resp.Ok {
		return CreateAppResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appManifestCreateMethod)
	}

	return resp.CreateAppResult, nil
}

type ExportAppResult struct {
	Manifest types.SlackYaml `json:"manifest,omitempty"`
}

type ExportAppResponse struct {
	extendedBaseResponse
	ExportAppResult
}

// ExportAppManifest calls "apps.manifest.export" to gather manifest details
func (c *Client) ExportAppManifest(ctx context.Context, token, appID string) (ExportAppResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsManifestExport")
	span.SetTag("app", appID)
	defer span.Finish()

	args := struct {
		AppID string `json:"app_id"`
	}{
		AppID: appID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return ExportAppResult{}, errInvalidArguments.WithRootCause(err)
	}
	b, err := c.postJSON(ctx, appManifestExportMethod, token, "", body)
	if err != nil {
		return ExportAppResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := ExportAppResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return ExportAppResult{}, errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appManifestExportMethod)
	}
	if !resp.Ok {
		return ExportAppResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appManifestExportMethod)
	}
	return resp.ExportAppResult, nil
}

type ValidateAppManifestResult struct {
	Warnings slackerror.Warnings
}

// ValidateAppManifest validates a new Slack app
func (c *Client) ValidateAppManifest(ctx context.Context, token string, manifest types.AppManifest, appId string) (ValidateAppManifestResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.ValidateAppManifest")
	defer span.Finish()

	// We send in an app_id if it exists so that the validation can be more robust - currently
	// it allows us to check for breaking changes since we have an existing app to compare
	// this one to. Note that this means it will be checking against whichever app version is
	// deployed (slack deploy), but cannot check against apps installed only locally (slack run)
	args := struct {
		Manifest types.AppManifest `json:"manifest,omitempty"`
		AppId    string            `json:"app_id,omitempty"`
	}{
		manifest,
		appId,
	}

	body, err := json.Marshal(args)

	if err != nil {
		return ValidateAppManifestResult{slackerror.Warnings{}},
			errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appManifestValidateMethod, token, "", body)

	if err != nil {
		return ValidateAppManifestResult{slackerror.Warnings{}},
			errHTTPRequestFailed.WithRootCause(err)
	}

	resp := extendedBaseResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return ValidateAppManifestResult{slackerror.Warnings{}},
			errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appManifestValidateMethod)
	}

	if resp.Ok && len(resp.Errors) == 0 {
		return ValidateAppManifestResult{resp.Warnings}, nil
	}

	return ValidateAppManifestResult{resp.Warnings},
		slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appManifestValidateMethod)
}

// UpdateAppResult details returned
type UpdateAppResult struct {
	AppID             string      `json:"app_id,omitempty"`
	Credentials       Credentials `json:"credentials,omitempty"` // but will undergo some slimming down in the near future (removal of credentials)
	OAuthAuthorizeURL string      `json:"oauth_authorize_url,omitempty"`
}

type updateAppResponse struct {
	extendedBaseResponse
	UpdateAppResult
}

// UpdateApp updates a Slack app
func (c *Client) UpdateApp(ctx context.Context, token string, appID string, manifest types.AppManifest, forceUpdate bool, continueWithBreakingChanges bool) (UpdateAppResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.UpdateApp")
	span.SetTag("app", appID)
	defer span.Finish()

	args := struct {
		Manifest               types.AppManifest `json:"manifest,omitempty"`
		AppID                  string            `json:"app_id,omitempty"`
		ForceUpdate            bool              `json:"force_update,omitempty"`
		ConsentBreakingChanges bool              `json:"consent_breaking_changes"`
	}{
		manifest,
		appID,
		forceUpdate,
		continueWithBreakingChanges,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return UpdateAppResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appManifestUpdateMethod, token, "", body)
	if err != nil {
		return UpdateAppResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := updateAppResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return UpdateAppResult{}, errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appManifestUpdateMethod)
	}

	if !resp.Ok {
		return UpdateAppResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appManifestUpdateMethod)
	}

	return resp.UpdateAppResult, nil
}

// GenerateS3PresignedPost details to be saved
type GenerateS3PresignedPostResult struct {
	Url      string              `json:"url"`
	FileName string              `json:"file_name"`
	Fields   PresignedPostFields `json:"fields"`
}
type generateS3PresignedPostResponse struct {
	extendedBaseResponse
	GenerateS3PresignedPostResult
}

func (c *Client) GetPresignedS3PostParams(ctx context.Context, token string, appID string) (GenerateS3PresignedPostResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.GetPresignedS3PostParams")
	span.SetTag("app", appID)
	defer span.Finish()

	args := struct {
		AppID string `json:"app_id,omitempty"`
	}{
		appID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return GenerateS3PresignedPostResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appGeneratePresignedPostMethod, token, "", body)
	if err != nil {
		return GenerateS3PresignedPostResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := generateS3PresignedPostResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return GenerateS3PresignedPostResult{}, errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appGeneratePresignedPostMethod)
	}

	if !resp.Ok {
		return GenerateS3PresignedPostResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appGeneratePresignedPostMethod)
	}

	return resp.GenerateS3PresignedPostResult, nil
}

type UploadAppArgs struct {
	AppID       string `json:"app_id,omitempty"`
	ArchiveFile string `json:"file,omitempty"`
}

type uploadAppResponse struct {
	extendedBaseResponse
}

// UploadApp takes a manifest and creates a new app
func (c *Client) UploadApp(ctx context.Context, token, runtime, appID string, fileName string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.UploadApp")
	span.SetTag("app", appID)
	defer span.Finish()

	args := struct {
		AppID    string `json:"app_id"`
		FileName string `json:"s3_key"`
		Runtime  string `json:"runtime"`
	}{
		appID,
		fileName,
		runtime,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appUploadMethod, token, "", body)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	resp := uploadAppResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appUploadMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appUploadMethod)
	}

	return nil
}

type deleteAppResponse struct {
	extendedBaseResponse
}

// DeleteApp fully deletes the Slack app identified by appID
func (c *Client) DeleteApp(ctx context.Context, token string, appID string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.DeleteApp")
	span.SetTag("app", appID)
	defer span.Finish()
	args := struct {
		AppID string `json:"app_id,omitempty"`
	}{
		appID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDeleteMethod, token, "", body)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	resp := deleteAppResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appDeleteMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDeleteMethod)
	}

	return nil
}

// UninstallApp uninstalls a Slack app from its team
func (c *Client) UninstallApp(ctx context.Context, token string, appID, teamID string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.UninstallApp")
	defer span.Finish()

	args := struct {
		AppID  string `json:"app_id,omitempty"`
		TeamID string `json:"team_id,omitempty"`
	}{
		appID,
		teamID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDeveloperUninstallMethod, token, "", body)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	resp := struct {
		baseResponse
	}{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appDeveloperUninstallMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, nil, appDeveloperUninstallMethod)
	}

	return nil
}

type AppStatusResultAppInfo struct {
	AppID            string                  `json:"app_id"`
	Installed        bool                    `json:"is_installed"`
	Hosted           bool                    `json:"is_hosted"`
	EnterpriseGrants []types.EnterpriseGrant `json:"enterprise_grants,omitempty"`
}

type GetAppStatusResult struct {
	Apps []AppStatusResultAppInfo `json:"apps"`
	Team struct {
		TeamID       string `json:"team_id"`
		TeamDomain   string `json:"team_domain"`
		IsEnterprise bool   `json:"is_enterprise"`
	} `json:"team"`
}

type GetAppStatusResponse struct {
	extendedBaseResponse
	GetAppStatusResult
}

// GetAppStatus fetches information about the given apps's installation
func (c *Client) GetAppStatus(ctx context.Context, token string, appIDs []string, teamID string) (GetAppStatusResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.GetAppStatus")
	defer span.Finish()

	args := struct {
		AppIDs []string `json:"app_ids,omitempty"`
		TeamID string   `json:"team_id,omitempty"`
	}{
		appIDs,
		teamID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return GetAppStatusResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appStatusMethod, token, "", body)
	if err != nil {
		return GetAppStatusResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	var resp GetAppStatusResponse
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return GetAppStatusResult{}, errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appStatusMethod)
	}

	if !resp.Ok {
		return GetAppStatusResult{}, slackerror.NewApiError(resp.Error, resp.Description, nil, appStatusMethod)
	}

	return resp.GetAppStatusResult, nil
}

type AppsApprovalsRequestsCreateResult struct {
	RequestID string `json:"request_id"`
}

type appsApprovalsRequestsCreateResponse struct {
	extendedBaseResponse
	AppsApprovalsRequestsCreateResult
}

type appsApprovalsRequestsCancelResponse struct {
	extendedBaseResponse
}

// GenerateS3PresignedPost details to be saved
type AppsConnectionsOpenResult struct {
	Url string `json:"url"`
}
type appsConnectionsOpenResponse struct {
	extendedBaseResponse
	AppsConnectionsOpenResult
}

func (c *Client) ConnectionsOpen(ctx context.Context, token string) (AppsConnectionsOpenResult, error) {
	args := struct{}{}

	body, err := json.Marshal(args)
	if err != nil {
		return AppsConnectionsOpenResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appConnectionsOpenMethod, token, "", body)
	if err != nil {
		return AppsConnectionsOpenResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := appsConnectionsOpenResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return AppsConnectionsOpenResult{}, errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appConnectionsOpenMethod)
	}

	if !resp.Ok {
		return AppsConnectionsOpenResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appConnectionsOpenMethod)
	}

	return resp.AppsConnectionsOpenResult, nil
}

// Details related to successful installation
type DeveloperAppInstallResult struct {
	AppID           string `json:"app_id,omitempty"`
	APIAccessTokens struct {
		Bot      string `json:"bot,omitempty"`
		AppLevel string `json:"app_level,omitempty"`
		User     string `json:"user,omitempty"`
	} `json:"api_access_tokens,omitempty"`
}

// Extended details returned when installation is denied due to AAA
type developerAppInstallAAARequiredResult struct {
	TeamID string `json:"team_id,omitempty"`
}

type developerAppInstallResponse struct {
	extendedBaseResponse
	developerAppInstallAAARequiredResult
	DeveloperAppInstallResult
}

func (c *Client) DeveloperAppInstall(ctx context.Context, IO iostreams.IOStreamer, token string, app types.App, botScopes []string, outgoingDomains []string, orgGrantWorkspaceID string, autoRequestAAAFlag bool) (DeveloperAppInstallResult, types.InstallState, error) {
	grantID := orgGrantWorkspaceID
	if grantID == types.GrantAllOrgWorkspaces && types.IsEnterpriseTeamID(app.TeamID) {
		// Passing in the enterprise ID will ensure grants are added for all workspaces in the org.
		grantID = app.EnterpriseID
	} else if !(types.IsEnterpriseTeamID(app.TeamID)) {
		// Should be omitted if the app is not being created on an org (standalone or enterprise-workspace created apps).
		grantID = ""
	}

	args := struct {
		AppID           string   `json:"app_id,omitempty"`
		BotScopes       []string `json:"bot_scopes,omitempty"`
		OutgoingDomains []string `json:"outgoing_domains,omitempty"`
		TeamID          string   `json:"team_id,omitempty"`
	}{
		app.AppID,
		botScopes,
		outgoingDomains,
		grantID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return DeveloperAppInstallResult{}, "", errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDeveloperInstallMethod, token, "", body)
	if err != nil {
		return DeveloperAppInstallResult{}, "", errHTTPRequestFailed.WithRootCause(err)
	}

	var resp developerAppInstallResponse
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return DeveloperAppInstallResult{}, "", errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appDeveloperInstallMethod)
	}

	if !resp.Ok {
		// Install requests to orgs (which have admin approval requirements for apps) may return an app_approval_* error message
		if goutils.Contains(appApprovalErrorMessages, resp.Error, false /*caseSensitive*/) {
			requestTeam := resp.TeamID
			if resp.TeamID == app.EnterpriseID {
				// apps.approvals.requests.create only accepts a workspace ID for team_id
				requestTeam = ""
			}
			var installState, err = c.handleAppApprovalStates(ctx, IO, resp.Error, token, app.AppID, requestTeam, strings.Join(botScopes, ","), outgoingDomains, autoRequestAAAFlag)
			return DeveloperAppInstallResult{}, installState, err
		}

		return DeveloperAppInstallResult{}, "", slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDeveloperInstallMethod)
	}

	return resp.DeveloperAppInstallResult, types.SUCCESS, nil
}

// handleAppApprovalStates handles responses from the developerInstall API when admin apps approval (AAA) is on
// AAA can be toggled on in standalone workspaces and is always on for Enterprise Grid organizations
func (c *Client) handleAppApprovalStates(ctx context.Context, IO iostreams.IOStreamer, resp, token, appID, teamID, scopes string, outgoingDomains []string, autoRequestAAA bool) (types.InstallState, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.handleAppApprovalStates")
	defer span.Finish()

	var alternativeSuggestion = "Alternatively, retry on a workspace without administrator approval turned on"
	if resp == slackerror.ErrAppApprovalRequestEligible {
		return c.handleAppRequestEligibleState(ctx, IO, resp, token, appID, teamID, scopes, outgoingDomains, alternativeSuggestion, autoRequestAAA)
	} else {
		// Ask the developer if they want to cancel their pending app approval request for this app
		if resp == slackerror.ErrAppApprovalRequestPending {
			return c.handleAppRequestPendingState(ctx, IO, resp, token, appID, teamID, scopes, outgoingDomains, alternativeSuggestion, autoRequestAAA)
		}

		return "", slackerror.New(resp).AppendRemediation(alternativeSuggestion)
	}
}

func (c *Client) handleAppRequestEligibleState(ctx context.Context, IO iostreams.IOStreamer, resp, token, appID, teamID, scopes string, outgoingDomains []string, alternativeSuggestion string, autoRequestAAA bool) (types.InstallState, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.handleAppRequestEligibleState")
	defer span.Finish()

	var administratorApprovalNotice = fmt.Sprintf("\n%s", style.Sectionf(style.TextSection{
		Emoji: "bell",
		Text:  "Administrator approval is required to install this app",
		Secondary: []string{
			alternativeSuggestion,
		},
	}))
	fmt.Println(administratorApprovalNotice)
	IO.PrintTrace(ctx, slacktrace.AdminAppApprovalRequestRequired)

	var shouldSendApprovalRequest bool
	var err error
	if autoRequestAAA {
		shouldSendApprovalRequest = true
	} else {
		// prompt as to whether to sent AAA request
		shouldSendApprovalRequest, err = c.io.ConfirmPrompt(ctx, "Request approval to install this app?", true)
		if err != nil {
			return "", err
		}
	}

	span.SetTag("app_approval_prompt_response", shouldSendApprovalRequest)
	span.SetTag("app_id", appID)

	// Create an app approval request
	var reason string
	if shouldSendApprovalRequest {
		IO.PrintTrace(ctx, slacktrace.AdminAppApprovalRequestShouldSend)
		if !autoRequestAAA {
			// prompt for reason
			reason, err = c.io.InputPrompt(ctx, "Enter a reason for installing this app:", iostreams.InputPromptConfig{
				Required: false,
			})
			if err != nil {
				return "", err
			}
		} else {
			reason = "This request has been automatically generated according to project environment settings."
		}
		IO.PrintTrace(ctx, slacktrace.AdminAppApprovalRequestReasonSubmitted, reason)

		_, err = c.RequestAppApproval(ctx, token, appID, teamID, reason, scopes, outgoingDomains)
		if err != nil {
			return "", err
		}

		IO.PrintTrace(ctx, slacktrace.AdminAppApprovalRequestPending)
		return types.REQUEST_PENDING, nil
	} else {
		return types.REQUEST_NOT_SENT, nil
	}
}

func (c *Client) handleAppRequestPendingState(ctx context.Context, IO iostreams.IOStreamer, resp, token, appID, teamID, scopes string, outgoingDomains []string, alternativeSuggestion string, autoRequestAAA bool) (types.InstallState, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.handleAppRequestPendingState")
	defer span.Finish()

	// Output current status
	var requestPendingErr = slackerror.New(resp)
	var status = fmt.Sprintf("\n%s", style.Sectionf(style.TextSection{
		Emoji: "bell",
		Text:  requestPendingErr.Message,
		Secondary: []string{
			requestPendingErr.Remediation,
		},
	}))
	fmt.Println(status)

	shouldCancelRequest, err := c.io.ConfirmPrompt(ctx, "Cancel the current request to install this app?", false)
	if err != nil {
		return "", err
	}

	span.SetTag("app_request_cancellation_prompt_response", shouldCancelRequest)
	span.SetTag("app_id", appID)

	// Create an app approval request
	if shouldCancelRequest {
		args := struct {
			AppID string `json:"app_id,omitempty"`
		}{
			appID,
		}

		body, err := json.Marshal(args)
		if err != nil {
			return "", errInvalidArguments.WithRootCause(err)
		}

		b, err := c.postJSON(ctx, appApprovalRequestCancelMethod, token, "", body)
		if err != nil {
			return "", errHTTPRequestFailed.WithRootCause(err)
		}

		appsApprovalsRequestsCancelResp := appsApprovalsRequestsCancelResponse{}
		err = goutils.JsonUnmarshal(b, &appsApprovalsRequestsCancelResp)
		if err != nil {
			return "", errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appApprovalRequestCancelMethod)
		}

		if !appsApprovalsRequestsCancelResp.Ok {
			return "", slackerror.NewApiError(
				appsApprovalsRequestsCancelResp.Error,
				appsApprovalsRequestsCancelResp.Description,
				appsApprovalsRequestsCancelResp.Errors,
				appApprovalRequestCancelMethod,
			)
		}

		// If the developer cancels their request, ask them if they want to send another request
		installState, err := c.handleAppRequestEligibleState(ctx, IO, resp, token, appID, teamID, scopes, outgoingDomains, alternativeSuggestion, autoRequestAAA)
		if err != nil {
			return "", err
		}

		if installState != types.REQUEST_PENDING {
			return types.REQUEST_CANCELLED, nil
		} else {
			return installState, nil
		}
	} else {
		return types.REQUEST_PENDING, nil
	}
}

// RequestAppApproval creates a new app approval request to slack
func (c *Client) RequestAppApproval(ctx context.Context, token string, appID string, teamID string, reason string, scopes string, outgoingDomains []string) (AppsApprovalsRequestsCreateResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "RequestAppApproval")
	defer span.Finish()

	args := struct {
		AppID   string   `json:"app,omitempty"`
		Scopes  string   `json:"bot_scopes,omitempty"`
		Reason  string   `json:"reason,omitempty"`
		Domains []string `json:"domains,omitempty"`
		TeamID  string   `json:"team_id,omitempty"`
	}{
		appID,
		scopes,
		reason,
		outgoingDomains,
		teamID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return AppsApprovalsRequestsCreateResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appApprovalRequestCreateMethod, token, "", body)
	if err != nil {
		return AppsApprovalsRequestsCreateResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := appsApprovalsRequestsCreateResponse{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return AppsApprovalsRequestsCreateResult{}, errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appApprovalRequestCreateMethod)
	}

	if !resp.Ok {
		return AppsApprovalsRequestsCreateResult{}, slackerror.NewApiError(
			resp.Error,
			resp.Description,
			resp.Errors,
			appApprovalRequestCreateMethod,
		)
	}

	return resp.AppsApprovalsRequestsCreateResult, nil
}
