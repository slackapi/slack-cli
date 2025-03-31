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

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/mock"
)

type ApiMock struct {
	mock.Mock
}

// WorkflowsClient

func (m *ApiMock) AddDefaultMocks() {
	m.On("Host").Return("https://slack.com")
}

func (m *ApiMock) WorkflowsTriggersCreate(ctx context.Context, token string, createRequest TriggerRequest) (types.DeployedTrigger, error) {
	args := m.Called(ctx, token, createRequest)
	return args.Get(0).(types.DeployedTrigger), args.Error(1)
}

func (m *ApiMock) WorkflowsTriggersUpdate(ctx context.Context, token string, updateRequest TriggerUpdateRequest) (types.DeployedTrigger, error) {
	args := m.Called(ctx, token, updateRequest)
	return args.Get(0).(types.DeployedTrigger), args.Error(1)
}

func (m *ApiMock) WorkflowsTriggersDelete(ctx context.Context, token string, triggerId string) error {
	args := m.Called(ctx, token, triggerId)
	return args.Error(0)
}

func (m *ApiMock) WorkflowsTriggersInfo(ctx context.Context, token string, triggerId string) (types.DeployedTrigger, error) {
	args := m.Called(ctx, token, triggerId)
	return args.Get(0).(types.DeployedTrigger), args.Error(1)
}

func (m *ApiMock) WorkflowsTriggersList(ctx context.Context, token string, listArgs TriggerListRequest) ([]types.DeployedTrigger, string, error) {
	args := m.Called(ctx, token, listArgs)
	return args.Get(0).([]types.DeployedTrigger), args.Get(1).(string), args.Error(2)
}

// SessionsClient

func (m *ApiMock) ValidateSession(ctx context.Context, token string) (AuthSession, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(AuthSession), args.Error(1)
}

func (m *ApiMock) RevokeToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// TriggerAccessClient

func (m *ApiMock) TriggerPermissionsList(ctx context.Context, token, triggerID string) (types.Permission, []string, error) {
	args := m.Called(ctx, token, triggerID)
	return args.Get(0).(types.Permission), args.Get(1).([]string), args.Error(2)
}

func (m *ApiMock) TriggerPermissionsSet(ctx context.Context, token, triggerID, entities string, distributionType types.Permission, entityType string) ([]string, error) {
	args := m.Called(ctx, token, triggerID, entities, distributionType, entityType)
	return args.Get(0).([]string), args.Error(1)
}

func (m *ApiMock) TriggerPermissionsAddEntities(ctx context.Context, token, triggerID, entities string, entityType string) error {
	args := m.Called(ctx, token, triggerID, entities, entityType)
	return args.Error(0)
}

func (m *ApiMock) TriggerPermissionsRemoveEntities(ctx context.Context, token, triggerID, entities string, entityType string) error {
	args := m.Called(ctx, token, triggerID, entities, entityType)
	return args.Error(0)
}

// Collaborator management

func (m *ApiMock) AddCollaborator(ctx context.Context, token, appID string, slackUser types.SlackUser) error {
	args := m.Called(ctx, token, appID, slackUser)
	return args.Error(0)
}

func (m *ApiMock) ListCollaborators(ctx context.Context, token, appID string) ([]types.SlackUser, error) {
	args := m.Called(ctx, token, appID)
	return args.Get(0).([]types.SlackUser), args.Error(1)
}

func (m *ApiMock) RemoveCollaborator(ctx context.Context, token, appID string, slackUser types.SlackUser) (slackerror.Warnings, error) {
	args := m.Called(ctx, token, appID, slackUser)
	return nil, args.Error(0)
}

func (m *ApiMock) UpdateCollaborator(ctx context.Context, token, appID string, slackUser types.SlackUser) error {
	args := m.Called(ctx, token, appID, slackUser)
	return args.Error(0)
}

// ActivityClient

func (m *ApiMock) Activity(ctx context.Context, token string, activityRequest types.ActivityRequest) (ActivityResult, error) {
	args := m.Called(ctx, token, activityRequest)
	return args.Get(0).(ActivityResult), args.Error(1)
}

// AuthClient

func (m *ApiMock) ExchangeAuthTicket(ctx context.Context, ticket string, challenge string, cliVersion string) (ExchangeAuthTicketResult, error) {
	args := m.Called(ctx, ticket, challenge, cliVersion)
	return args.Get(0).(ExchangeAuthTicketResult), args.Error(1)
}

func (m *ApiMock) GenerateAuthTicket(ctx context.Context, cliVersion string, serviceTokenFlag bool) (GenerateAuthTicketResult, error) {
	args := m.Called(ctx, cliVersion, serviceTokenFlag)
	return args.Get(0).(GenerateAuthTicketResult), args.Error(1)
}

func (m *ApiMock) RotateToken(ctx context.Context, auth types.SlackAuth) (RotateTokenResult, error) {
	args := m.Called(ctx, auth)
	return args.Get(0).(RotateTokenResult), args.Error(1)
}

// UserClient

func (m *ApiMock) UsersInfo(ctx context.Context, token, userID string) (*types.UserInfo, error) {
	args := m.Called(ctx, token, userID)
	return args.Get(0).(*types.UserInfo), args.Error(1)
}

// ChannelClient

func (m *ApiMock) ChannelsInfo(ctx context.Context, token, channelID string) (*types.ChannelInfo, error) {
	args := m.Called(ctx, token, channelID)
	return args.Get(0).(*types.ChannelInfo), args.Error(1)
}

// TeamClient (team and organization share the same client)

func (m *ApiMock) TeamsInfo(ctx context.Context, token, teamID string) (*types.TeamInfo, error) {
	args := m.Called(ctx, token, teamID)
	return args.Get(0).(*types.TeamInfo), args.Error(1)
}

func (m *ApiMock) AuthTeamsList(ctx context.Context, token string, limit int) ([]types.TeamInfo, string, error) {
	args := m.Called(ctx, token)
	return args.Get(0).([]types.TeamInfo), args.String(1), args.Error(2)
}

// ExternalAuthClient

func (m *ApiMock) AppsAuthExternalStart(ctx context.Context, token, appID, providerKey string) (string, error) {
	args := m.Called(ctx, token, appID, providerKey)
	return args.Get(0).(string), args.Error(1)
}

func (m *ApiMock) AppsAuthExternalDelete(ctx context.Context, token, appID, providerKey string, externalTokenId string) error {
	args := m.Called(ctx, token, appID, providerKey, externalTokenId)
	return args.Error(0)
}

func (m *ApiMock) AppsAuthExternalList(ctx context.Context, token, appID string, includeWorkflows bool) (types.ExternalAuthorizationInfoLists, error) {
	args := m.Called(ctx, token, appID)
	return args.Get(0).(types.ExternalAuthorizationInfoLists), args.Error(1)
}

func (m *ApiMock) AppsAuthExternalClientSecretAdd(ctx context.Context, token, appID, providerKey, clientSecret string) error {
	args := m.Called(ctx, token, appID, providerKey, clientSecret)
	return args.Error(0)
}

func (m *ApiMock) AppsAuthExternalSelectAuth(ctx context.Context, token, appID, providerKey, workflowId, externalTokenId string) error {
	args := m.Called(ctx, token, appID, providerKey, workflowId, externalTokenId)
	return args.Error(0)
}

// FunctionDistributionClient

func (m *ApiMock) FunctionDistributionList(ctx context.Context, callbackID, appID string) (types.Permission, []types.FunctionDistributionUser, error) {
	args := m.Called(ctx, callbackID, appID)
	return args.Get(0).(types.Permission), args.Get(1).([]types.FunctionDistributionUser), args.Error(2)
}

func (m *ApiMock) FunctionDistributionSet(ctx context.Context, callbackID, appID string, distributionType types.Permission, users string) ([]types.FunctionDistributionUser, error) {
	args := m.Called(ctx, callbackID, appID, distributionType, users)
	return args.Get(0).([]types.FunctionDistributionUser), args.Error(1)
}

func (m *ApiMock) FunctionDistributionAddUsers(ctx context.Context, callbackID, appID, users string) error {
	args := m.Called(ctx, callbackID, appID, users)
	return args.Error(0)
}

func (m *ApiMock) FunctionDistributionRemoveUsers(ctx context.Context, callbackID, appID, users string) error {
	args := m.Called(ctx, callbackID, appID, users)
	return args.Error(0)
}

// DatastoresClient

func (m *ApiMock) AppsDatastorePut(ctx context.Context, token string, request types.AppDatastorePut) (types.AppDatastorePutResult, error) {
	args := m.Called(ctx, token, request)
	return args.Get(0).(types.AppDatastorePutResult), args.Error(1)
}

func (m *ApiMock) AppsDatastoreBulkPut(ctx context.Context, token string, request types.AppDatastoreBulkPut) (types.AppDatastoreBulkPutResult, error) {
	args := m.Called(ctx, token, request)
	return args.Get(0).(types.AppDatastoreBulkPutResult), args.Error(1)
}

func (m *ApiMock) AppsDatastoreUpdate(ctx context.Context, token string, request types.AppDatastoreUpdate) (types.AppDatastoreUpdateResult, error) {
	args := m.Called(ctx, token, request)
	return args.Get(0).(types.AppDatastoreUpdateResult), args.Error(1)
}

func (m *ApiMock) AppsDatastoreGet(ctx context.Context, token string, request types.AppDatastoreGet) (types.AppDatastoreGetResult, error) {
	args := m.Called(ctx, token, request)
	return args.Get(0).(types.AppDatastoreGetResult), args.Error(1)
}

func (m *ApiMock) AppsDatastoreBulkGet(ctx context.Context, token string, request types.AppDatastoreBulkGet) (types.AppDatastoreBulkGetResult, error) {
	args := m.Called(ctx, token, request)
	return args.Get(0).(types.AppDatastoreBulkGetResult), args.Error(1)
}

func (m *ApiMock) AppsDatastoreDelete(ctx context.Context, token string, request types.AppDatastoreDelete) (types.AppDatastoreDeleteResult, error) {
	args := m.Called(ctx, token, request)
	return args.Get(0).(types.AppDatastoreDeleteResult), args.Error(1)
}

func (m *ApiMock) AppsDatastoreBulkDelete(ctx context.Context, token string, request types.AppDatastoreBulkDelete) (types.AppDatastoreBulkDeleteResult, error) {
	args := m.Called(ctx, token, request)
	return args.Get(0).(types.AppDatastoreBulkDeleteResult), args.Error(1)
}

func (m *ApiMock) AppsDatastoreQuery(ctx context.Context, token string, query types.AppDatastoreQuery) (types.AppDatastoreQueryResult, error) {
	args := m.Called(ctx, token, query)
	return args.Get(0).(types.AppDatastoreQueryResult), args.Error(1)
}

func (m *ApiMock) AppsDatastoreCount(ctx context.Context, token string, query types.AppDatastoreCount) (types.AppDatastoreCountResult, error) {
	args := m.Called(ctx, token, query)
	return args.Get(0).(types.AppDatastoreCountResult), args.Error(1)
}

// StepsClient

func (m *ApiMock) StepsList(ctx context.Context, token string, workflow string, appId string) ([]StepVersion, error) {
	args := m.Called(ctx, token, workflow, appId)
	return args.Get(0).([]StepVersion), args.Error(1)
}

func (m *ApiMock) StepsResponsesExport(ctx context.Context, token string, workflow string, appId string, stepId string) error {
	args := m.Called(ctx, token, workflow, appId, stepId)
	return args.Error(0)
}

// AppsClient

func (m *ApiMock) DeleteApp(ctx context.Context, token string, appID string) error {
	args := m.Called(ctx, token, appID)
	return args.Error(0)
}

func (m *ApiMock) UninstallApp(ctx context.Context, token string, appID, teamID string) error {
	args := m.Called(ctx, token, appID, teamID)
	return args.Error(0)
}

func (m *ApiMock) GetAppStatus(ctx context.Context, token string, appIDs []string, teamID string) (GetAppStatusResult, error) {
	args := m.Called(ctx, token, appIDs, teamID)
	return args.Get(0).(GetAppStatusResult), args.Error(1)
}

func (m *ApiMock) SetHost(host string) {
	m.Called(host)
}

func (m *ApiMock) CertifiedAppInstall(ctx context.Context, token string, certifiedAppId string) (CertifiedInstallResult, error) {
	args := m.Called(ctx, token, certifiedAppId)
	return args.Get(0).(CertifiedInstallResult), args.Error(1)
}

func (m *ApiMock) RequestAppApproval(ctx context.Context, token string, appID string, teamID string, reason string, scopes string, outgoingDomains []string) (AppsApprovalsRequestsCreateResult, error) {
	args := m.Called(ctx, token, appID, teamID, reason, scopes, outgoingDomains)
	return args.Get(0).(AppsApprovalsRequestsCreateResult), args.Error(1)
}

// VariablesClient

func (m *ApiMock) AddVariable(ctx context.Context, token, appID, name, value string) error {
	args := m.Called(ctx, token, appID, name, value)
	return args.Error(0)
}

func (m *ApiMock) ListVariables(ctx context.Context, token, appID string) ([]string, error) {
	args := m.Called(ctx, token, appID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *ApiMock) RemoveVariable(ctx context.Context, token string, appID string, variableName string) error {
	args := m.Called(ctx, token, appID, variableName)
	return args.Error(0)
}

func (m *ApiMock) Host() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *ApiMock) ConnectionsOpen(ctx context.Context, token string) (AppsConnectionsOpenResult, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(AppsConnectionsOpenResult), args.Error(1)
}

func (m *ApiMock) ExportAppManifest(ctx context.Context, token string, appId string) (ExportAppResult, error) {
	args := m.Called(ctx, token, appId)
	return args.Get(0).(ExportAppResult), args.Error(1)
}

func (m *ApiMock) ValidateAppManifest(ctx context.Context, token string, manifest types.AppManifest, appId string) (ValidateAppManifestResult, error) {
	args := m.Called(ctx, token, manifest, appId)
	return args.Get(0).(ValidateAppManifestResult), args.Error(1)
}

func (m *ApiMock) CreateApp(ctx context.Context, token string, manifest types.AppManifest, enableDistribution bool) (CreateAppResult, error) {
	args := m.Called(ctx, token, manifest, enableDistribution)
	return args.Get(0).(CreateAppResult), args.Error(1)
}

func (m *ApiMock) UpdateApp(ctx context.Context, token string, appID string, manifest types.AppManifest, forceUpdate bool, continueWithBreakingChanges bool) (UpdateAppResult, error) {
	args := m.Called(ctx, token, appID, manifest, forceUpdate, continueWithBreakingChanges)
	return args.Get(0).(UpdateAppResult), args.Error(1)
}

func (m *ApiMock) GetPresignedS3PostParams(ctx context.Context, token string, appID string) (GenerateS3PresignedPostResult, error) {
	args := m.Called(ctx, token, appID)
	return args.Get(0).(GenerateS3PresignedPostResult), args.Error(1)
}

func (m *ApiMock) UploadApp(ctx context.Context, token, runtime, appID string, fileName string) error {
	args := m.Called(ctx, token, runtime, appID, fileName)
	return args.Error(0)
}

func (m *ApiMock) DeveloperAppInstall(ctx context.Context, IO iostreams.IOStreamer, token string, app types.App, botScopes []string, outgoingDomains []string, orgGrantWorkspaceID string, autoAAARequest bool) (DeveloperAppInstallResult, types.InstallState, error) {
	args := m.Called(ctx, IO, token, app, botScopes, outgoingDomains, orgGrantWorkspaceID, autoAAARequest)
	return args.Get(0).(DeveloperAppInstallResult), args.Get(1).(types.InstallState), args.Error(2)
}
